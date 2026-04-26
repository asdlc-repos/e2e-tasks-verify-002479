package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"tasks-api/internal/db"
	"tasks-api/internal/models"
)

// Handler holds the dependencies for HTTP handlers
type Handler struct {
	db *db.DB
}

// New creates a new Handler with the given database
func New(database *db.DB) *Handler {
	return &Handler{db: database}
}

// writeJSON writes a JSON response with the given status code and data
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// TasksHandler handles GET /tasks and POST /tasks
func (h *Handler) TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTasks(w, r)
	case http.MethodPost:
		h.createTask(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// TaskByIDHandler handles PATCH /tasks/{id} and DELETE /tasks/{id}
func (h *Handler) TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path: /tasks/{id}
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id := strings.TrimSuffix(path, "/")

	if id == "" {
		writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		h.updateTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listTasks handles GET /tasks
func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.db.GetAllTasks()
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve tasks")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// createTask handles POST /tasks
func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate and trim description
	description := strings.TrimSpace(req.Description)
	if description == "" {
		writeError(w, http.StatusBadRequest, "description is required and cannot be empty")
		return
	}

	task, err := h.db.CreateTask(description)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

// updateTask handles PATCH /tasks/{id}
func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, id string) {
	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.db.UpdateTask(id, req.Completed)
	if err != nil {
		log.Printf("Error updating task %s: %v", id, err)
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// deleteTask handles DELETE /tasks/{id}
func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, id string) {
	deleted, err := h.db.DeleteTask(id)
	if err != nil {
		log.Printf("Error deleting task %s: %v", id, err)
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	if !deleted {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
