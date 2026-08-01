package service

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
)

func mustSetUserPin(t *testing.T, userId, channelId, ttlSeconds int) {
	t.Helper()
	if err := SetUserPin(userId, channelId, ttlSeconds); err != nil {
		t.Fatalf("SetUserPin(%d, %d, %d) failed: %v", userId, channelId, ttlSeconds, err)
	}
}

func pinsForUser(t *testing.T, userId int) []UserChannelPin {
	t.Helper()
	pins, err := ListUserPins()
	if err != nil {
		t.Fatalf("ListUserPins failed: %v", err)
	}
	out := make([]UserChannelPin, 0, 1)
	for _, p := range pins {
		if p.UserId == userId {
			out = append(out, p)
		}
	}
	return out
}

func newTestGinContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}

func TestUserChannelPinSetGet(t *testing.T) {
	const userId, channelId = 910001, 42
	t.Cleanup(func() { ClearUserPin(userId) })
	mustSetUserPin(t, userId, channelId, 60)
	got, ok := GetUserPin(userId)
	if !ok || got != channelId {
		t.Fatalf("GetUserPin = (%d, %v), want (%d, true)", got, ok, channelId)
	}
}

func TestUserChannelPinGetMiss(t *testing.T) {
	if got, ok := GetUserPin(910002); ok {
		t.Fatalf("GetUserPin on missing pin = (%d, true), want miss", got)
	}
}

func TestUserChannelPinInvalidParams(t *testing.T) {
	cases := [][3]int{{0, 1, 60}, {1, 0, 60}, {1, 1, 0}, {-1, 5, 60}, {5, -1, 60}, {5, 1, -3}}
	for _, tc := range cases {
		if err := SetUserPin(tc[0], tc[1], tc[2]); err == nil {
			t.Fatalf("SetUserPin(%d, %d, %d) expected error, got nil", tc[0], tc[1], tc[2])
		}
	}
}

func TestUserChannelPinOverwrite(t *testing.T) {
	const userId = 910003
	t.Cleanup(func() { ClearUserPin(userId) })
	mustSetUserPin(t, userId, 11, 60)
	mustSetUserPin(t, userId, 22, 120)
	got, ok := GetUserPin(userId)
	if !ok || got != 22 {
		t.Fatalf("GetUserPin after overwrite = (%d, %v), want (22, true)", got, ok)
	}
	pins := pinsForUser(t, userId)
	if len(pins) != 1 || pins[0].ChannelId != 22 {
		t.Fatalf("list after overwrite = %+v, want single entry with channel 22", pins)
	}
}

func TestUserChannelPinClear(t *testing.T) {
	const userId = 910004
	mustSetUserPin(t, userId, 7, 60)
	ClearUserPin(userId)
	if got, ok := GetUserPin(userId); ok {
		t.Fatalf("GetUserPin after clear = (%d, true), want miss", got)
	}
	if pins := pinsForUser(t, userId); len(pins) != 0 {
		t.Fatalf("list after clear = %+v, want empty", pins)
	}
}

func TestUserChannelPinTTLExpire(t *testing.T) {
	const userId = 910005
	t.Cleanup(func() { ClearUserPin(userId) })
	mustSetUserPin(t, userId, 9, 1)
	time.Sleep(1200 * time.Millisecond)
	if got, ok := GetUserPin(userId); ok {
		t.Fatalf("GetUserPin after ttl expiry = (%d, true), want miss", got)
	}
	if pins := pinsForUser(t, userId); len(pins) != 0 {
		t.Fatalf("expired pin still listed: %+v", pins)
	}
}

func TestUserChannelPinList(t *testing.T) {
	users := map[int]int{910006: 1, 910007: 2, 910008: 3}
	for uid, ch := range users {
		mustSetUserPin(t, uid, ch, 60)
		uid := uid
		t.Cleanup(func() { ClearUserPin(uid) })
	}
	pins, err := ListUserPins()
	if err != nil {
		t.Fatalf("ListUserPins failed: %v", err)
	}
	found := 0
	for _, p := range pins {
		want, ok := users[p.UserId]
		if !ok {
			continue
		}
		if p.ChannelId != want {
			t.Fatalf("pin %+v, want channel %d", p, want)
		}
		if p.ExpiresAt <= time.Now().Unix() {
			t.Fatalf("pin %+v has non-future expires_at", p)
		}
		found++
	}
	if found != len(users) {
		t.Fatalf("found %d pins, want %d; all pins: %+v", found, len(users), pins)
	}
}

func TestUserChannelPinExpiryHelpers(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if got := pinExpiresAt(now, 60); got != now.Unix()+60 {
		t.Fatalf("pinExpiresAt = %d, want %d", got, now.Unix()+60)
	}
	cases := []struct {
		expiresAt int64
		want      bool
	}{
		{0, false}, // unknown expiry is treated as not expired
		{now.Unix() + 1, false},
		{now.Unix(), true},
		{now.Unix() - 1, true},
	}
	for _, tc := range cases {
		if got := isPinExpired(tc.expiresAt, now); got != tc.want {
			t.Fatalf("isPinExpired(%d) = %v, want %v", tc.expiresAt, got, tc.want)
		}
	}
}

func TestApplyUserChannelPin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const userId, channelId = 910009, 77
	t.Cleanup(func() { ClearUserPin(userId) })
	mustSetUserPin(t, userId, channelId, 60)

	// Pin hit: specific_channel_id injected as string.
	c := newTestGinContext()
	c.Set("id", userId)
	ApplyUserChannelPin(c)
	v, ok := common.GetContextKey(c, constant.ContextKeyTokenSpecificChannelId)
	if !ok {
		t.Fatal("specific_channel_id not injected for pinned user")
	}
	s, isStr := v.(string)
	if !isStr || s != strconv.Itoa(channelId) {
		t.Fatalf("injected value = %#v, want string %q", v, strconv.Itoa(channelId))
	}

	// Existing specific_channel_id (admin token suffix) must not be overwritten.
	c2 := newTestGinContext()
	c2.Set("id", userId)
	common.SetContextKey(c2, constant.ContextKeyTokenSpecificChannelId, "5")
	ApplyUserChannelPin(c2)
	if v2, _ := common.GetContextKey(c2, constant.ContextKeyTokenSpecificChannelId); v2 != "5" {
		t.Fatalf("existing specific_channel_id overwritten: %#v", v2)
	}

	// Missing user id: nothing injected.
	c3 := newTestGinContext()
	ApplyUserChannelPin(c3)
	if _, ok := common.GetContextKey(c3, constant.ContextKeyTokenSpecificChannelId); ok {
		t.Fatal("specific_channel_id injected without user id")
	}
}
