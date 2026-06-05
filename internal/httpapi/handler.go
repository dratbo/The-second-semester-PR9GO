package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"example.com/pz9-redis-cache/internal/service"
	"example.com/pz9-redis-cache/internal/task"
)

type Handler struct {
	service *service.TaskService
}

func NewHandler(service *service.TaskService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var page, limit int
	if pageStr != "" || limitStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}
	}

	tasks, err := h.service.GetTasksList(r.Context(), page, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	t, err := h.service.GetTaskByID(r.Context(), id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(t)
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var t task.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTask(r.Context(), t); err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
