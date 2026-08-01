package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/pkg/cachex"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/samber/hot"
)

const (
	userChannelPinNamespace  = "new-api:user_channel_pin:v1"
	userChannelPinListPrefix = "new-api:user_channel_pin:list:"

	userChannelPinCacheCapacity     = 100_000
	userChannelPinDefaultTTLSeconds = 3600

	// UserChannelPinMaxTTLSeconds is the maximum ttl allowed for a user channel pin (7 days).
	UserChannelPinMaxTTLSeconds = 7 * 24 * 3600

	userChannelPinRedisOpTimeout   = 2 * time.Second
	userChannelPinRedisScanTimeout = 30 * time.Second
)

var (
	userChannelPinCacheOnce sync.Once
	userChannelPinCache     *cachex.HybridCache[int]

	// userChannelPinMemIndex is the list index used when Redis is unavailable,
	// storing map[int]UserChannelPin with lazy eviction of expired entries on read.
	userChannelPinMemIndex sync.Map
)

// UserChannelPin describes one active user-level channel pin.
type UserChannelPin struct {
	UserId    int   `json:"user_id"`
	ChannelId int   `json:"channel_id"`
	ExpiresAt int64 `json:"expires_at"`
}

func getUserChannelPinCache() *cachex.HybridCache[int] {
	userChannelPinCacheOnce.Do(func() {
		userChannelPinCache = cachex.NewHybridCache[int](cachex.HybridCacheConfig[int]{
			Namespace: cachex.Namespace(userChannelPinNamespace),
			Redis:     common.RDB,
			RedisEnabled: func() bool {
				return common.RedisEnabled && common.RDB != nil
			},
			RedisCodec: cachex.IntCodec{},
			Memory: func() *hot.HotCache[string, int] {
				return hot.NewHotCache[string, int](hot.LRU, userChannelPinCacheCapacity).
					WithTTL(time.Duration(userChannelPinDefaultTTLSeconds) * time.Second).
					WithJanitor().
					Build()
			},
		})
	})
	return userChannelPinCache
}

func userChannelPinRedisOn() bool {
	return common.RedisEnabled && common.RDB != nil
}

func userChannelPinListKey(userId int) string {
	return userChannelPinListPrefix + strconv.Itoa(userId)
}

// pinExpiresAt computes the unix timestamp at which a pin created at now with ttlSeconds expires.
func pinExpiresAt(now time.Time, ttlSeconds int) int64 {
	return now.Add(time.Duration(ttlSeconds) * time.Second).Unix()
}

// isPinExpired reports whether a pin with the given expiresAt unix timestamp is expired at now.
// expiresAt <= 0 means the expiry is unknown and the pin is treated as not expired.
func isPinExpired(expiresAt int64, now time.Time) bool {
	return expiresAt > 0 && !now.Before(time.Unix(expiresAt, 0))
}

// SetUserPin pins userId to channelId for ttlSeconds, so that all relay requests from
// this user are routed to the pinned channel until the pin expires or is cleared.
func SetUserPin(userId int, channelId int, ttlSeconds int) error {
	if userId <= 0 || channelId <= 0 || ttlSeconds <= 0 {
		return fmt.Errorf("invalid user channel pin params: user_id=%d, channel_id=%d, ttl_seconds=%d",
			userId, channelId, ttlSeconds)
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	if err := getUserChannelPinCache().SetWithTTL(strconv.Itoa(userId), channelId, ttl); err != nil {
		return err
	}
	return setUserPinIndex(userId, channelId, ttlSeconds, time.Now())
}

// setUserPinIndex maintains the list index alongside the main pin cache.
func setUserPinIndex(userId int, channelId int, ttlSeconds int, now time.Time) error {
	if userChannelPinRedisOn() {
		ctx, cancel := context.WithTimeout(context.Background(), userChannelPinRedisOpTimeout)
		defer cancel()
		ttl := time.Duration(ttlSeconds) * time.Second
		return common.RDB.Set(ctx, userChannelPinListKey(userId), channelId, ttl).Err()
	}
	userChannelPinMemIndex.Store(userId, UserChannelPin{
		UserId:    userId,
		ChannelId: channelId,
		ExpiresAt: pinExpiresAt(now, ttlSeconds),
	})
	return nil
}

// ClearUserPin removes the pin for userId from both the pin cache and the list index.
func ClearUserPin(userId int) {
	if userId <= 0 {
		return
	}
	if _, err := getUserChannelPinCache().DeleteMany([]string{strconv.Itoa(userId)}); err != nil {
		common.SysError("clear user channel pin cache failed: " + err.Error())
	}
	if userChannelPinRedisOn() {
		ctx, cancel := context.WithTimeout(context.Background(), userChannelPinRedisOpTimeout)
		defer cancel()
		if err := common.RDB.Del(ctx, userChannelPinListKey(userId)).Err(); err != nil {
			common.SysError("clear user channel pin index failed: " + err.Error())
		}
		return
	}
	userChannelPinMemIndex.Delete(userId)
}

// GetUserPin returns the pinned channel id for userId, if a valid pin exists.
func GetUserPin(userId int) (int, bool) {
	if userId <= 0 {
		return 0, false
	}
	channelId, found, err := getUserChannelPinCache().Get(strconv.Itoa(userId))
	if err != nil {
		common.SysError("get user channel pin failed: " + err.Error())
		return 0, false
	}
	if !found || channelId <= 0 {
		return 0, false
	}
	return channelId, true
}

// ListUserPins returns all active user channel pins, sorted by user id.
func ListUserPins() ([]UserChannelPin, error) {
	if userChannelPinRedisOn() {
		return listUserPinsRedis()
	}
	return listUserPinsMemory(time.Now()), nil
}

func listUserPinsRedis() ([]UserChannelPin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), userChannelPinRedisScanTimeout)
	defer cancel()

	pins := make([]UserChannelPin, 0, 16)
	var cursor uint64
	for {
		keys, next, err := common.RDB.Scan(ctx, cursor, userChannelPinListPrefix+"*", 1000).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			if pin, ok := loadUserPinFromRedis(ctx, key); ok {
				pins = append(pins, pin)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	sortUserPins(pins)
	return pins, nil
}

func loadUserPinFromRedis(ctx context.Context, key string) (UserChannelPin, bool) {
	userId, err := strconv.Atoi(strings.TrimPrefix(key, userChannelPinListPrefix))
	if err != nil || userId <= 0 {
		return UserChannelPin{}, false
	}
	val, err := common.RDB.Get(ctx, key).Result()
	if err != nil {
		// redis.Nil means the key expired between SCAN and GET, which is fine.
		if !errors.Is(err, redis.Nil) {
			common.SysError("load user channel pin index failed: " + err.Error())
		}
		return UserChannelPin{}, false
	}
	channelId, err := strconv.Atoi(val)
	if err != nil || channelId <= 0 {
		return UserChannelPin{}, false
	}
	pin := UserChannelPin{UserId: userId, ChannelId: channelId}
	if ttl, err := common.RDB.TTL(ctx, key).Result(); err == nil && ttl > 0 {
		pin.ExpiresAt = time.Now().Add(ttl).Unix()
	}
	return pin, true
}

func listUserPinsMemory(now time.Time) []UserChannelPin {
	pins := make([]UserChannelPin, 0, 16)
	userChannelPinMemIndex.Range(func(k, v any) bool {
		pin, ok := v.(UserChannelPin)
		if !ok {
			return true
		}
		if isPinExpired(pin.ExpiresAt, now) {
			userChannelPinMemIndex.Delete(k)
			return true
		}
		pins = append(pins, pin)
		return true
	})
	sortUserPins(pins)
	return pins
}

func sortUserPins(pins []UserChannelPin) {
	sort.Slice(pins, func(i, j int) bool { return pins[i].UserId < pins[j].UserId })
}

// ApplyUserChannelPin injects the pinned channel id into the request context using the
// same context key consumed by middleware.Distribute, so a pinned user behaves exactly
// like an admin token with an explicit channel suffix. An explicit specific_channel_id
// already set during token auth always wins and is never overwritten.
func ApplyUserChannelPin(c *gin.Context) {
	if _, ok := common.GetContextKey(c, constant.ContextKeyTokenSpecificChannelId); ok {
		return
	}
	userId := c.GetInt("id")
	if userId <= 0 {
		return
	}
	channelId, ok := GetUserPin(userId)
	if !ok {
		return
	}
	common.SetContextKey(c, constant.ContextKeyTokenSpecificChannelId, strconv.Itoa(channelId))
}
