package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"LinkStorageService/internal/service"
)

type LinkHandler struct {
	service *service.LinkService
}

func NewLinkHandler(service *service.LinkService) *LinkHandler {
	return &LinkHandler{service: service}
}

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	shortCode, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"short_code": shortCode,
	})
}

func (h *LinkHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		writeError(w, http.StatusBadRequest, "short_code is required")
		return
	}

	link, err := h.service.GetByCodeAndIncrement(r.Context(), shortCode)
	if err != nil {
		if err.Error() == "link not found" {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"url":    link.OriginalURL,
		"visits": link.Visits,
	})
}

func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	links, total, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]map[string]interface{}, len(links))
	for i, link := range links {
		items[i] = map[string]interface{}{
			"short_code": link.ShortCode,
			"url":        link.OriginalURL,
			"visits":     link.Visits,
			"created_at": link.CreatedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		writeError(w, http.StatusBadRequest, "short_code is required")
		return
	}

	if err := h.service.Delete(r.Context(), shortCode); err != nil {
		if err.Error() == "link not found" {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LinkHandler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		writeError(w, http.StatusBadRequest, "short_code is required")
		return
	}

	link, err := h.service.GetStats(r.Context(), shortCode)
	if err != nil {
		if err.Error() == "link not found" {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"short_code": link.ShortCode,
		"url":        link.OriginalURL,
		"visits":     link.Visits,
		"created_at": link.CreatedAt.Format(time.RFC3339),
	})
}

func (h *LinkHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
