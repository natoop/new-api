package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/affinity"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/auth"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/config"
)

type Server struct {
	config config.Config
	store  *affinity.Store
	auth   auth.AdminAuthenticator
	logger *slog.Logger
}

func New(config config.Config, store *affinity.Store, authenticator auth.AdminAuthenticator, logger *slog.Logger) *Server {
	return &Server{config: config, store: store, auth: authenticator, logger: logger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/v1/admin/channel-affinities", s.bindings)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) bindings(w http.ResponseWriter, r *http.Request) {
	remoteIP, ok := requestIP(r)
	if !ok || !s.config.AllowsIP(remoteIP) {
		writeError(w, http.StatusForbidden, "SOURCE_NOT_ALLOWED", "request source is not allowed")
		return
	}
	authorization := r.Header.Get("Authorization")
	if _, ok := bearerToken(authorization); !ok {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "ADMIN_AUTH_REQUIRED", "administrator authentication is required")
		return
	}
	authContext, cancel := context.WithTimeout(r.Context(), s.config.AuthTimeout)
	actor, err := s.auth.Authenticate(authContext, authorization)
	cancel()
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthenticated):
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "ADMIN_AUTH_REQUIRED", "new-api did not accept the supplied token")
		case errors.Is(err, auth.ErrNotAdmin):
			writeError(w, http.StatusForbidden, "ADMIN_ROLE_REQUIRED", "new-api administrator role is required")
		default:
			s.logger.Error("new-api administrator verification failed", "error", err)
			writeError(w, http.StatusServiceUnavailable, "NEW_API_AUTH_UNAVAILABLE", "could not verify administrator identity")
		}
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.upsert(w, r, actor, remoteIP)
	case http.MethodGet:
		s.get(w, r)
	case http.MethodDelete:
		s.delete(w, r, actor, remoteIP)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method is not allowed")
	}
}

func (s *Server) upsert(w http.ResponseWriter, r *http.Request, actor auth.Actor, remoteIP net.IP) {
	var binding affinity.Binding
	if err := decodeJSON(r, &binding); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if err := affinity.Validate(binding, s.config.AllowAutoGroup, true); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BINDING", err.Error())
		return
	}
	requestID := requestID(r)
	ctx, cancel := context.WithTimeout(r.Context(), s.config.RedisTimeout)
	defer cancel()
	err := s.store.Upsert(ctx, binding, s.audit("upsert", binding, actor, requestID, remoteIP))
	if err != nil {
		s.logger.Error("channel affinity upsert failed", "request_id", requestID, "user_id", binding.UserID, "group", binding.Group, "model", binding.Model, "channel_id", binding.ChannelID, "error", err)
		writeError(w, http.StatusServiceUnavailable, "REDIS_WRITE_FAILED", "could not persist the affinity binding")
		return
	}
	s.logger.Info("channel affinity upserted", "request_id", requestID, "user_id", binding.UserID, "group", binding.Group, "model", binding.Model, "channel_id", binding.ChannelID)
	writeJSON(w, http.StatusOK, bindingResponse{Binding: binding, TTLSeconds: int(s.config.TTL.Seconds()), RequestID: requestID})
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	binding, err := bindingFromQuery(r, s.config.AllowAutoGroup, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BINDING", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), s.config.RedisTimeout)
	defer cancel()
	lookup, found, err := s.store.Get(ctx, binding)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "REDIS_READ_FAILED", "could not read the affinity binding")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "BINDING_NOT_FOUND", "no affinity binding exists for this user, group, and model")
		return
	}
	writeJSON(w, http.StatusOK, bindingResponse{Binding: lookup.Binding, TTLSeconds: int(lookup.TTL.Seconds()), RequestID: requestID(r)})
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request, actor auth.Actor, remoteIP net.IP) {
	binding, err := bindingFromQuery(r, s.config.AllowAutoGroup, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BINDING", err.Error())
		return
	}
	requestID := requestID(r)
	ctx, cancel := context.WithTimeout(r.Context(), s.config.RedisTimeout)
	defer cancel()
	deleted, err := s.store.Delete(ctx, binding, s.audit("delete", binding, actor, requestID, remoteIP))
	if err != nil {
		s.logger.Error("channel affinity delete failed", "request_id", requestID, "user_id", binding.UserID, "group", binding.Group, "model", binding.Model, "error", err)
		writeError(w, http.StatusServiceUnavailable, "REDIS_DELETE_FAILED", "could not remove the affinity binding")
		return
	}
	s.logger.Info("channel affinity deleted", "request_id", requestID, "user_id", binding.UserID, "group", binding.Group, "model", binding.Model, "deleted", deleted)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted, "request_id": requestID})
}

func (s *Server) audit(action string, binding affinity.Binding, actor auth.Actor, requestID string, remoteIP net.IP) affinity.AuditEvent {
	return affinity.AuditEvent{
		Action:    action,
		Binding:   binding,
		ActorHint: actorHint(actor),
		RequestID: requestID,
		RemoteIP:  remoteIP.String(),
		Occurred:  time.Now(),
	}
}

type bindingResponse struct {
	affinity.Binding
	TTLSeconds int    `json:"ttl_seconds"`
	RequestID  string `json:"request_id"`
}

func bindingFromQuery(r *http.Request, allowAutoGroup bool, requireChannel bool) (affinity.Binding, error) {
	query := r.URL.Query()
	userID, err := parsePositiveInt(query.Get("user_id"))
	if err != nil {
		return affinity.Binding{}, errors.New("user_id must be a positive integer")
	}
	channelID := 0
	if requireChannel {
		channelID, err = parsePositiveInt(query.Get("channel_id"))
		if err != nil {
			return affinity.Binding{}, errors.New("channel_id must be a positive integer")
		}
	}
	binding := affinity.Binding{UserID: userID, Group: query.Get("group"), Model: query.Get("model"), ChannelID: channelID}
	if err := affinity.Validate(binding, allowAutoGroup, requireChannel); err != nil {
		return affinity.Binding{}, err
	}
	return binding, nil
}

func decodeJSON(r *http.Request, destination any) error {
	if r.Body == nil {
		return errors.New("JSON request body is required")
	}
	defer r.Body.Close()
	if err := common.DecodeJson(io.LimitReader(r.Body, 16<<10), destination); err != nil {
		return errors.New("request body must be valid JSON")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	payload, err := common.Marshal(value)
	if err != nil {
		http.Error(w, "response encoding failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func bearerToken(raw string) (string, bool) {
	parts := strings.Fields(raw)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func requestIP(r *http.Request) (net.IP, bool) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, false
	}
	ip := net.ParseIP(host)
	return ip, ip != nil
}

func actorHint(actor auth.Actor) string {
	if actor.Username == "" {
		return "user:" + strconv.Itoa(actor.UserID)
	}
	return "user:" + strconv.Itoa(actor.UserID) + ":" + actor.Username
}

func requestID(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if value == "" || len(value) > 128 || strings.ContainsAny(value, "\r\n") {
		return "-"
	}
	return value
}

func parsePositiveInt(raw string) (int, error) {
	if raw == "" {
		return 0, errors.New("not numeric")
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New("not numeric")
	}
	if value <= 0 {
		return 0, errors.New("not positive")
	}
	return value, nil
}
