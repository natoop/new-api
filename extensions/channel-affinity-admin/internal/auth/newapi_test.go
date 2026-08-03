package auth

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthenticateDelegatesRoleDecisionToNewAPI(t *testing.T) {
	baseURL, err := url.Parse("http://new-api.internal")
	require.NoError(t, err)
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, "/api/user/self", request.URL.Path)
		require.Equal(t, "Bearer existing-token", request.Header.Get("Authorization"))
		return jsonResponse(`{"success":true,"data":{"id":99,"role":10,"username":"admin"}}`), nil
	})}

	actor, err := NewNewAPIAuthenticator(baseURL, client).Authenticate(context.Background(), "Bearer existing-token")

	require.NoError(t, err)
	require.Equal(t, Actor{UserID: 99, Role: 10, Username: "admin"}, actor)
}

func TestAuthenticateRejectsCommonUser(t *testing.T) {
	baseURL, err := url.Parse("http://new-api.internal")
	require.NoError(t, err)
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(`{"success":true,"data":{"id":99,"role":1}}`), nil
	})}

	_, err = NewNewAPIAuthenticator(baseURL, client).Authenticate(context.Background(), "Bearer existing-token")

	require.ErrorIs(t, err, ErrNotAdmin)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
