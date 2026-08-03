package affinity

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-redis/redis/v8"
)

const Namespace = "new-api:channel_affinity:v1"

type Binding struct {
	UserID    int    `json:"user_id"`
	Group     string `json:"group"`
	Model     string `json:"model"`
	ChannelID int    `json:"channel_id"`
}

type Lookup struct {
	Binding
	TTL time.Duration `json:"-"`
}

type AuditEvent struct {
	Action    string
	Binding   Binding
	ActorHint string
	RequestID string
	RemoteIP  string
	Occurred  time.Time
}

type Store struct {
	rdb         redis.UniversalClient
	ruleName    string
	ttl         time.Duration
	auditStream string
}

func NewStore(rdb redis.UniversalClient, ruleName string, ttl time.Duration, auditStream string) *Store {
	return &Store{rdb: rdb, ruleName: ruleName, ttl: ttl, auditStream: auditStream}
}

func (s *Store) Upsert(ctx context.Context, binding Binding, event AuditEvent) error {
	key := s.Key(binding)
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, key, strconv.Itoa(binding.ChannelID), s.ttl)
		pipe.XAdd(ctx, &redis.XAddArgs{Stream: s.auditStream, Values: auditValues(event)})
		return nil
	})
	if err != nil {
		return fmt.Errorf("write affinity binding and audit event: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, binding Binding) (Lookup, bool, error) {
	key := s.Key(binding)
	value, err := s.rdb.Get(ctx, key).Result()
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
	ttl, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return Lookup{}, false, fmt.Errorf("read affinity binding TTL: %w", err)
	}
	return Lookup{Binding: Binding{UserID: binding.UserID, Group: binding.Group, Model: binding.Model, ChannelID: channelID}, TTL: ttl}, true, nil
}

func (s *Store) Delete(ctx context.Context, binding Binding, event AuditEvent) (bool, error) {
	key := s.Key(binding)
	commands, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, key)
		pipe.XAdd(ctx, &redis.XAddArgs{Stream: s.auditStream, Values: auditValues(event)})
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

// Key deliberately mirrors new-api's buildChannelAffinityCacheKeySuffix:
// <namespace>:<rule>:<model>:<group>:<user-id>.
func (s *Store) Key(binding Binding) string {
	return strings.Join([]string{Namespace, s.ruleName, binding.Model, binding.Group, strconv.Itoa(binding.UserID)}, ":")
}

func Validate(binding Binding, allowAutoGroup bool, requireChannel bool) error {
	if binding.UserID <= 0 {
		return errors.New("user_id must be a positive integer")
	}
	if requireChannel && binding.ChannelID <= 0 {
		return errors.New("channel_id must be a positive integer")
	}
	if err := validateKeyPart("group", binding.Group); err != nil {
		return err
	}
	if !allowAutoGroup && binding.Group == "auto" {
		return errors.New("group=auto is disabled; provide an actual group name")
	}
	if err := validateKeyPart("model", binding.Model); err != nil {
		return err
	}
	return nil
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

func auditValues(event AuditEvent) map[string]any {
	return map[string]any{
		"action":      event.Action,
		"user_id":     event.Binding.UserID,
		"group":       event.Binding.Group,
		"model":       event.Binding.Model,
		"channel_id":  event.Binding.ChannelID,
		"actor_hint":  event.ActorHint,
		"request_id":  event.RequestID,
		"remote_ip":   event.RemoteIP,
		"occurred_at": event.Occurred.UTC().Format(time.RFC3339Nano),
	}
}
