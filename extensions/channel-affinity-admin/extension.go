package channelaffinityadmin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/affinity"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

// Register adds the isolated administrator-only affinity management API.
func Register(apiRouter *gin.RouterGroup) {
	route := apiRouter.Group("/extensions/channel-affinity")
	route.Use(middleware.AdminAuth())
	route.PUT("", upsert)
	route.GET("", get)
	route.DELETE("", deleteBinding)
}

func upsert(c *gin.Context) {
	var binding affinity.Binding
	if err := common.DecodeJson(c.Request.Body, &binding); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "request body must be valid JSON"})
		return
	}
	requestID := requestID(c)
	ttl, err := affinity.Upsert(c.Request.Context(), binding, auditEvent(c, "upsert", binding, requestID))
	if err != nil {
		common.SysError("channel affinity admin upsert failed: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"user_id": binding.UserID, "group": binding.Group, "model": binding.Model, "channel_id": binding.ChannelID, "ttl_seconds": int(ttl.Seconds()), "request_id": requestID}})
}

func get(c *gin.Context) {
	binding, err := queryBinding(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	lookup, found, err := affinity.Get(c.Request.Context(), binding)
	if err != nil {
		common.SysError("channel affinity admin get failed: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "affinity binding not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"user_id": lookup.UserID, "group": lookup.Group, "model": lookup.Model, "channel_id": lookup.ChannelID, "ttl_seconds": int(lookup.TTL.Seconds()), "request_id": requestID(c)}})
}

func deleteBinding(c *gin.Context) {
	binding, err := queryBinding(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	requestID := requestID(c)
	deleted, err := affinity.Delete(c.Request.Context(), binding, auditEvent(c, "delete", binding, requestID))
	if err != nil {
		common.SysError("channel affinity admin delete failed: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"deleted": deleted, "request_id": requestID}})
}

func queryBinding(c *gin.Context) (affinity.Binding, error) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		return affinity.Binding{}, errors.New("user_id must be a positive integer")
	}
	binding := affinity.Binding{UserID: userID, Group: c.Query("group"), Model: c.Query("model")}
	if err := affinity.Validate(binding, false); err != nil {
		return affinity.Binding{}, err
	}
	return binding, nil
}

func auditEvent(c *gin.Context, action string, binding affinity.Binding, requestID string) affinity.AuditEvent {
	return affinity.AuditEvent{Action: action, Binding: binding, ActorID: c.GetInt("id"), RequestID: requestID, RemoteIP: c.ClientIP()}
}

func requestID(c *gin.Context) string {
	value := strings.TrimSpace(c.GetHeader("X-Request-ID"))
	if value == "" || len(value) > 128 || strings.ContainsAny(value, "\r\n") {
		return "-"
	}
	return value
}
