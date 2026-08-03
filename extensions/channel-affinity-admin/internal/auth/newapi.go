package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

const AdminRole = 10

var (
	ErrUnauthenticated = errors.New("new-api did not accept the supplied token")
	ErrNotAdmin        = errors.New("new-api user is not an administrator")
)

type Actor struct {
	UserID   int
	Role     int
	Username string
}

type AdminAuthenticator interface {
	Authenticate(ctx context.Context, authorization string) (Actor, error)
}

// NewAPIAuthenticator delegates authorization to the same /api/user/self
// endpoint that Switcher uses. It deliberately has no local user/role cache:
// role changes and disabled users take effect on the next management request.
type NewAPIAuthenticator struct {
	baseURL *url.URL
	client  *http.Client
}

func NewNewAPIAuthenticator(baseURL *url.URL, client *http.Client) *NewAPIAuthenticator {
	if client == nil {
		client = http.DefaultClient
	}
	return &NewAPIAuthenticator{baseURL: baseURL, client: client}
}

func (a *NewAPIAuthenticator) Authenticate(ctx context.Context, authorization string) (Actor, error) {
	if a == nil || a.baseURL == nil || strings.TrimSpace(authorization) == "" {
		return Actor{}, ErrUnauthenticated
	}
	endpoint := *a.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/user/self"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Actor{}, fmt.Errorf("create new-api identity request: %w", err)
	}
	request.Header.Set("Authorization", authorization)
	request.Header.Set("Accept", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		return Actor{}, fmt.Errorf("call new-api identity endpoint: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return Actor{}, ErrUnauthenticated
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Actor{}, fmt.Errorf("new-api identity endpoint returned HTTP %d", response.StatusCode)
	}

	var envelope selfEnvelope
	if err := common.DecodeJson(io.LimitReader(response.Body, 64<<10), &envelope); err != nil {
		return Actor{}, fmt.Errorf("decode new-api identity response: %w", err)
	}
	if !envelope.Success || envelope.Data.ID <= 0 {
		return Actor{}, ErrUnauthenticated
	}
	if envelope.Data.Role < AdminRole {
		return Actor{}, ErrNotAdmin
	}
	return Actor{UserID: envelope.Data.ID, Role: envelope.Data.Role, Username: strings.TrimSpace(envelope.Data.Username)}, nil
}

type selfEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		ID       int    `json:"id"`
		Role     int    `json:"role"`
		Username string `json:"username"`
	} `json:"data"`
}
