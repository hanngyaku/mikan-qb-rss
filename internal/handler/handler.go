package handler

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/example/mikan-qb-rss/internal/config"
	"github.com/example/mikan-qb-rss/internal/model"
	"github.com/example/mikan-qb-rss/internal/qbittorrent"
	"github.com/example/mikan-qb-rss/internal/service"
)

type Handler struct {
	db      *sql.DB
	subs    *service.SubscriptionService
	logPath string
}

func New(db *sql.DB, logPath string) *Handler {
	return &Handler{db: db, subs: service.NewSubscriptionService(db), logPath: logPath}
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
	mux.HandleFunc("PUT /api/subscriptions/{id}", h.updateSubscription)
	mux.HandleFunc("DELETE /api/subscriptions/{id}", h.deleteSubscription)
	mux.HandleFunc("POST /api/subscriptions/{id}/sync", h.syncSubscription)
	mux.HandleFunc("GET /api/logs", h.getLogs)
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
// @Summary 添加订阅并同步 qBittorrent
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

// updateSubscription godoc
// @Summary 更新订阅并同步 qBittorrent
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "订阅 ID"
// @Param body body model.UpdateSubscriptionRequest true "订阅"
// @Success 200 {object} model.Subscription
// @Router /subscriptions/{id} [put]
func (h *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := subscriptionID(r)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	var req model.UpdateSubscriptionRequest
	if err := decode(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	item, err := h.subs.Update(r.Context(), id, req)
	if err != nil {
		handleSubscriptionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// deleteSubscription godoc
// @Summary 删除订阅及 qBittorrent RSS 配置
// @Tags subscriptions
// @Param id path int true "订阅 ID"
// @Success 204
// @Router /subscriptions/{id} [delete]
func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := subscriptionID(r)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.subs.Delete(r.Context(), id); err != nil {
		handleSubscriptionError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// syncSubscription godoc
// @Summary 重新同步订阅到 qBittorrent
// @Tags subscriptions
// @Param id path int true "订阅 ID"
// @Success 200 {object} model.Subscription
// @Router /subscriptions/{id}/sync [post]
func (h *Handler) syncSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := subscriptionID(r)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	item, err := h.subs.Sync(r.Context(), id)
	if err != nil {
		handleSubscriptionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// getLogs godoc
// @Summary 获取最新日志
// @Tags logs
// @Produce json
// @Param lines query int false "返回行数，默认 100，最大 1000"
// @Success 200 {object} model.LogResponse
// @Router /logs [get]
func (h *Handler) getLogs(w http.ResponseWriter, r *http.Request) {
	count := 100
	if value := r.URL.Query().Get("lines"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 1000 {
			fail(w, http.StatusBadRequest, errors.New("lines must be between 1 and 1000"))
			return
		}
		count = parsed
	}
	file, err := os.Open(h.logPath)
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()
	lines := make([]string, 0, count)
	next := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(lines) < count {
			lines = append(lines, scanner.Text())
		} else {
			lines[next] = scanner.Text()
			next = (next + 1) % count
		}
	}
	if err := scanner.Err(); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	if len(lines) == count && next > 0 {
		lines = append(append([]string{}, lines[next:]...), lines[:next]...)
	}
	writeJSON(w, http.StatusOK, model.LogResponse{Lines: lines})
}

func subscriptionID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid subscription ID")
	}
	return id, nil
}

func handleSubscriptionError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, http.StatusNotFound, errors.New("subscription not found"))
		return
	}
	fail(w, http.StatusBadGateway, err)
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
