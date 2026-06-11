package middleware

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func playgroundAbort(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "new_api_error",
			"code":    "playground_token_required",
		},
	})
	c.Abort()
}

// PlaygroundTokenAuth enforces that the playground runs against the caller's
// OWN API token (created on the Tokens page) — it must NOT silently consume
// quota through the session. Runs after UserAuth and BEFORE Distribute so the
// user sees the token prompt instead of a downstream channel error.
func PlaygroundTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.Header.Get("X-Playground-Token")
		if strings.HasPrefix(key, "Bearer ") || strings.HasPrefix(key, "bearer ") {
			key = strings.TrimSpace(key[7:])
		}
		key = strings.TrimPrefix(key, "sk-")
		if idx := strings.IndexByte(key, '-'); idx >= 0 {
			key = key[:idx]
		}
		if key == "" {
			playgroundAbort(c, "请在游乐场填写你自己的 API 令牌（在「令牌」页创建）/ Please enter your own API token (create one on the Tokens page)")
			return
		}
		token, err := model.ValidateUserToken(key)
		if err != nil || token == nil {
			playgroundAbort(c, "API 令牌无效或已禁用 / API token is invalid or disabled")
			return
		}
		if token.UserId != c.GetInt("id") {
			playgroundAbort(c, "只能使用你自己的 API 令牌 / You can only use your own API token")
			return
		}
		// The caller's real token drives quota / model / group limits downstream.
		if err := SetupContextForToken(c, token); err != nil {
			playgroundAbort(c, "API 令牌无效或已禁用 / API token is invalid or disabled")
			return
		}
		c.Next()
	}
}
