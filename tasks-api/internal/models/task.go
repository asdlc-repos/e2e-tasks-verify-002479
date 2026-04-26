package models

import "time"

// Task represents a todo task
type Task struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateTaskRequest is the request body for creating a task
type CreateTaskRequest struct {
	Description string `json:"description"`
}

// UpdateTaskRequest is the request body for updating a task's completion status
type UpdateTaskRequest struct {
	Completed bool `json:"completed"`
}
