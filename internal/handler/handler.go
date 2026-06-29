package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/example/mikan-qb-rss/internal/config"
	"github.com/example/mikan-qb-rss/internal/model"
	"github.com/example/mikan-qb-rss/internal/qbittorrent"
	"github.com/example/mikan-qb-rss/internal/service"
)

type Handler struct {
	db   *sql.DB
	subs *service.SubscriptionService
}

func New(db *sql.DB) *Handler {
	return &Handler{db: db, subs: service.NewSubscriptionService(db)}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/settings", h.getSettings)
	mux.HandleFunc("PUT /api/settings", h.putSettings)
	mux.HandleFunc("POST /api/qb/test", h.testQB)
	mux.HandleFunc("GET /api/subscriptions", h.listSubscriptions)
	mux.HandleFunc("POST /api/subscriptions", h.createSubscription)
}

// getSettings godoc
// @Summary 获取设置（不返回密码）
// @Tags settings
// @Produce json
// @Success 200 {object} model.SettingsResponse
// @Router /settings [get]
func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := config.Get(r.Context(), h.db)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, config.Public(settings))
}

// putSettings godoc
// @Summary 更新设置，空密码保留原密码
// @Tags settings
// @Accept json
// @Produce json
// @Param body body model.UpdateSettingsRequest true "设置"
// @Success 200 {object} model.SettingsResponse
// @Router /settings [put]
func (h *Handler) putSettings(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateSettingsRequest
	if err := decode(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if !validHTTPURL(req.QBURL) || strings.TrimSpace(req.DownloadRoot) == "" || strings.TrimSpace(req.DefaultCategory) == "" || req.RSSInterval < 1 {
		fail(w, http.StatusBadRequest, errors.New("invalid settings"))
		return
	}
	if err := config.Update(r.Context(), h.db, req); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	settings, _ := config.Get(r.Context(), h.db)
	writeJSON(w, http.StatusOK, config.Public(settings))
}

// testQB godoc
// @Summary 测试 qBittorrent 连接
// @Tags qbittorrent
// @Produce json
// @Success 200 {object} model.QBTestResponse
// @Router /qb/test [post]
func (h *Handler) testQB(w http.ResponseWriter, r *http.Request) {
	settings, err := config.Get(r.Context(), h.db)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	client, err := qbittorrent.New(settings.QBURL, settings.QBUsername, settings.QBPassword)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	version, apiVersion, err := client.Test(r.Context())
	if err != nil {
		fail(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, model.QBTestResponse{Connected: true, Version: version, WebAPIVersion: apiVersion})
}

// listSubscriptions godoc
// @Summary 订阅列表
// @Tags subscriptions
// @Produce json
// @Success 200 {array} model.Subscription
// @Router /subscriptions [get]
func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	items, err := h.subs.List(r.Context())
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// createSubscription godoc
// @Summary 添加订阅（第一阶段仅写入本地数据库）
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param body body model.CreateSubscriptionRequest true "订阅"
// @Success 201 {object} model.Subscription
// @Router /subscriptions [post]
func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := decode(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	item, err := h.subs.Create(r.Context(), req)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func decode(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func fail(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func validHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}
