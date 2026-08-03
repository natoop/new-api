package affinity

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/go-redis/redis/v8"
)

const (
	Namespace = "new-api:channel_affinity:v1"
	RuleName  = "external-admin-user-channel-v1"

	auditStream = "new-api:channel_affinity_admin:v1:audit"
)

type Binding struct {
	UserID    int    `json:"user_id"`
	Group     string `json:"group"`
	Model     string `json:"model"`
	ChannelID int    `json:"channel_id"`
}

type Lookup struct {
	Binding
	TTL time.Duration
}

type AuditEvent struct {
	Action    string
	Binding   Binding
	ActorID   int
	RequestID string
	RemoteIP  string
}

func Upsert(ctx context.Context, binding Binding, event AuditEvent) (time.Duration, error) {
	ttl, err := configuredTTL()
	if err != nil {
		return 0, err
	}
	if err := Validate(binding, true); err != nil {
		return 0, err
	}
	key := Key(binding)
	_, err = common.RDB.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, key, strconv.Itoa(binding.ChannelID), ttl)
		pipe.XAdd(ctx, &redis.XAddArgs{Stream: auditStream, Values: auditValues(event)})
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("write affinity binding and audit event: %w", err)
	}
	return ttl, nil
}

func Get(ctx context.Context, binding Binding) (Lookup, bool, error) {
	if _, err := configuredTTL(); err != nil {
		return Lookup{}, false, err
	}
	if err := Validate(binding, false); err != nil {
		return Lookup{}, false, err
	}
	key := Key(binding)
	value, err := common.RDB.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return Lookup{}, false, nil
	}
	if err != nil {
		return Lookup{}, false, fmt.Errorf("read affinity binding: %w", err)
	}
	channelID, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || channelID <= 0 {
		return Lookup{}, false, fmt.Errorf("invalid channel ID stored at %q", key)
	}
	ttl, err := common.RDB.TTL(ctx, key).Result()
	if err != nil {
		return Lookup{}, false, fmt.Errorf("read affinity binding TTL: %w", err)
	}
	return Lookup{Binding: Binding{UserID: binding.UserID, Group: binding.Group, Model: binding.Model, ChannelID: channelID}, TTL: ttl}, true, nil
}

func Delete(ctx context.Context, binding Binding, event AuditEvent) (bool, error) {
	if _, err := configuredTTL(); err != nil {
		return false, err
	}
	if err := Validate(binding, false); err != nil {
		return false, err
	}
	commands, err := common.RDB.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, Key(binding))
		pipe.XAdd(ctx, &redis.XAddArgs{Stream: auditStream, Values: auditValues(event)})
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("delete affinity binding and write audit event: %w", err)
	}
	if len(commands) == 0 {
		return false, errors.New("Redis transaction returned no commands")
	}
	deleted, err := commands[0].(*redis.IntCmd).Result()
	if err != nil {
		return false, fmt.Errorf("delete affinity binding: %w", err)
	}
	return deleted > 0, nil
}

// Key mirrors new-api's channel-affinity cache protocol.
func Key(binding Binding) string {
	return strings.Join([]string{Namespace, RuleName, binding.Model, binding.Group, strconv.Itoa(binding.UserID)}, ":")
}

func Validate(binding Binding, requireChannel bool) error {
	if binding.UserID <= 0 {
		return errors.New("user_id must be a positive integer")
	}
	if requireChannel && binding.ChannelID <= 0 {
		return errors.New("channel_id must be a positive integer")
	}
	if binding.Group == "auto" {
		return errors.New("group=auto is not supported; provide an actual group name")
	}
	if err := validateKeyPart("group", binding.Group); err != nil {
		return err
	}
	return validateKeyPart("model", binding.Model)
}

func configuredTTL() (time.Duration, error) {
	if !common.RedisEnabled || common.RDB == nil {
		return 0, errors.New("Redis is not enabled")
	}
	setting := operation_setting.GetChannelAffinitySetting()
	if setting == nil || !setting.Enabled {
		return 0, errors.New("channel affinity is not enabled")
	}
	for _, rule := range setting.Rules {
		if rule.Name != RuleName {
			continue
		}
		if !rule.IncludeRuleName || !rule.IncludeModelName || !rule.IncludeUsingGroup || !isUserIDRule(rule) {
			return 0, errors.New("external channel-affinity rule does not match the documented contract")
		}
		ttlSeconds := rule.TTLSeconds
		if ttlSeconds <= 0 {
			ttlSeconds = setting.DefaultTTLSeconds
		}
		if ttlSeconds <= 0 {
			ttlSeconds = 3600
		}
		return time.Duration(ttlSeconds) * time.Second, nil
	}
	return 0, errors.New("external channel-affinity rule is not configured")
}

func isUserIDRule(rule operation_setting.ChannelAffinityRule) bool {
	return len(rule.KeySources) == 1 &&
		rule.KeySources[0].Type == "context_int" &&
		rule.KeySources[0].Key == "id" &&
		rule.ValueRegex == ""
}

func validateKeyPart(name, value string) error {
	if value == "" || !utf8.ValidString(value) {
		return fmt.Errorf("%s must be a non-empty UTF-8 string", name)
	}
	if strings.ContainsAny(value, ":\r\n\x00") {
		return fmt.Errorf("%s must not contain colon, newline, or NUL", name)
	}
	if len(value) > 256 {
		return fmt.Errorf("%s must be at most 256 bytes", name)
	}
	return nil
}

func auditValues(event AuditEvent) map[string]interface{} {
	return map[string]interface{}{
		"action":      event.Action,
		"user_id":     event.Binding.UserID,
		"group":       event.Binding.Group,
		"model":       event.Binding.Model,
		"channel_id":  event.Binding.ChannelID,
		"actor_id":    event.ActorID,
		"request_id":  event.RequestID,
		"remote_ip":   event.RemoteIP,
		"occurred_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
}
