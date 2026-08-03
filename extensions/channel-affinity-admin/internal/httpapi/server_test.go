package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/auth"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAdminBindingEndpointRejectsRequestsWithoutDedicatedAdminCredential(t *testing.T) {
	server := New(config.Config{}, nil, stubAuthenticator{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	request := httptest.NewRequest(http.MethodGet, "/v1/admin/channel-affinities?user_id=1&group=vip&model=gpt-4.1", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Equal(t, "Bearer", response.Header().Get("WWW-Authenticate"))
}

type stubAuthenticator struct{}

func (stubAuthenticator) Authenticate(context.Context, string) (auth.Actor, error) {
	return auth.Actor{UserID: 1, Role: auth.AdminRole}, nil
}
