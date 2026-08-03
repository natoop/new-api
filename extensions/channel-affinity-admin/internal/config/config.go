package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	defaultListenAddr = "127.0.0.1:8089"
	defaultRuleName   = "external-admin-user-channel-v1"
	defaultAudit      = "new-api:channel_affinity_admin:v1:audit"
)

// Config intentionally contains only the two external contracts: Redis and a
// dedicated administrator credential. It does not import or call new-api.
type Config struct {
	ListenAddr     string
	RedisOptions   *redis.Options
	NewAPIBaseURL  *url.URL
	AuthTimeout    time.Duration
	RuleName       string
	TTL            time.Duration
	AllowAutoGroup bool
	AllowedCIDRs   []*net.IPNet
	AuditStream    string
	RedisTimeout   time.Duration
}

func Load() (Config, error) {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		return Config{}, errors.New("REDIS_URL is required")
	}
	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		return Config{}, fmt.Errorf("parse REDIS_URL: %w", err)
	}

	allowedCIDRs, err := parseCIDRs(os.Getenv("AFFINITY_ALLOWED_CIDRS"))
	if err != nil {
		return Config{}, err
	}

	ttlSeconds, err := positiveIntEnv("AFFINITY_TTL_SECONDS", 3600, 60, 2_592_000)
	if err != nil {
		return Config{}, err
	}
	timeoutSeconds, err := positiveIntEnv("AFFINITY_REDIS_TIMEOUT_SECONDS", 3, 1, 30)
	if err != nil {
		return Config{}, err
	}
	authTimeoutSeconds, err := positiveIntEnv("AFFINITY_NEW_API_AUTH_TIMEOUT_SECONDS", 5, 1, 30)
	if err != nil {
		return Config{}, err
	}

	newAPIBaseURL, err := parseNewAPIBaseURL(os.Getenv("NEW_API_BASE_URL"))
	if err != nil {
		return Config{}, err
	}

	ruleName := strings.TrimSpace(os.Getenv("AFFINITY_RULE_NAME"))
	if ruleName == "" {
		ruleName = defaultRuleName
	}
	if err := validateKeyPart("AFFINITY_RULE_NAME", ruleName); err != nil {
		return Config{}, err
	}

	listenAddr := strings.TrimSpace(os.Getenv("AFFINITY_LISTEN_ADDR"))
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}
	if _, _, err := net.SplitHostPort(listenAddr); err != nil {
		return Config{}, fmt.Errorf("invalid AFFINITY_LISTEN_ADDR: %w", err)
	}

	auditStream := strings.TrimSpace(os.Getenv("AFFINITY_AUDIT_STREAM"))
	if auditStream == "" {
		auditStream = defaultAudit
	}

	return Config{
		ListenAddr:     listenAddr,
		RedisOptions:   redisOptions,
		NewAPIBaseURL:  newAPIBaseURL,
		AuthTimeout:    time.Duration(authTimeoutSeconds) * time.Second,
		RuleName:       ruleName,
		TTL:            time.Duration(ttlSeconds) * time.Second,
		AllowAutoGroup: strings.EqualFold(strings.TrimSpace(os.Getenv("AFFINITY_ALLOW_AUTO_GROUP")), "true"),
		AllowedCIDRs:   allowedCIDRs,
		AuditStream:    auditStream,
		RedisTimeout:   time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

func (c Config) AllowsIP(ip net.IP) bool {
	if len(c.AllowedCIDRs) == 0 {
		return true
	}
	for _, cidr := range c.AllowedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func parseCIDRs(raw string) ([]*net.IPNet, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	result := make([]*net.IPNet, 0, len(parts))
	for _, part := range parts {
		_, cidr, err := net.ParseCIDR(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid AFFINITY_ALLOWED_CIDRS value %q: %w", part, err)
		}
		result = append(result, cidr)
	}
	return result, nil
}

func parseNewAPIBaseURL(raw string) (*url.URL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, errors.New("NEW_API_BASE_URL is required")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("NEW_API_BASE_URL must be an absolute HTTP(S) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("NEW_API_BASE_URL must use http or https")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed, nil
}

func positiveIntEnv(name string, fallback, min, max int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", name, min, max)
	}
	return value, nil
}

func validateKeyPart(name, value string) error {
	if strings.ContainsAny(value, ":\r\n\x00") {
		return fmt.Errorf("%s must not contain colon, newline, or NUL", name)
	}
	return nil
}
