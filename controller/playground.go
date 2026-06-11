package controller

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func Playground(c *gin.Context) {
	var newAPIError *types.NewAPIError

	defer func() {
		if newAPIError != nil {
			c.JSON(newAPIError.StatusCode, gin.H{
				"error": newAPIError.ToOpenAIError(),
			})
		}
	}()

	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		newAPIError = types.NewError(errors.New("暂不支持使用 access token"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
		return
	}

	_, err := relaycommon.GenRelayInfo(c, types.RelayFormatOpenAI, nil, nil)
	if err != nil {
		newAPIError = types.NewError(err, types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
		return
	}

	userId := c.GetInt("id")

	// Playground must run against the caller's OWN API token (created on the
	// Tokens page) — it must NOT silently consume quota through the session.
	// The frontend supplies the key via the X-Playground-Token header.
	key := c.Request.Header.Get("X-Playground-Token")
	if strings.HasPrefix(key, "Bearer ") || strings.HasPrefix(key, "bearer ") {
		key = strings.TrimSpace(key[7:])
	}
	key = strings.TrimPrefix(key, "sk-")
	if idx := strings.IndexByte(key, '-'); idx >= 0 {
		key = key[:idx]
	}
	if key == "" {
		newAPIError = types.NewError(errors.New("请在游乐场填写你自己的 API 令牌（在「令牌」页创建）/ Please enter your own API token (create one on the Tokens page)"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
		return
	}
	token, err := model.ValidateUserToken(key)
	if err != nil || token == nil {
		newAPIError = types.NewError(errors.New("API 令牌无效或已禁用 / API token is invalid or disabled"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
		return
	}
	if token.UserId != userId {
		newAPIError = types.NewError(errors.New("只能使用你自己的 API 令牌 / You can only use your own API token"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
		return
	}

	// Write user context to ensure acceptUnsetRatio is available
	userCache, err := model.GetUserCache(userId)
	if err != nil {
		newAPIError = types.NewError(err, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
		return
	}
	userCache.WriteContext(c)

	// Use the caller's real token (its own quota / model / group limits apply).
	_ = middleware.SetupContextForToken(c, token)

	Relay(c, types.RelayFormatOpenAI)
}
