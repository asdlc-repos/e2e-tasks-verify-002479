package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"tasks-api/internal/models"
)

// DB wraps the sql.DB with a mutex for concurrent access safety
type DB struct {
	conn *sql.DB
	mu   sync.Mutex
}

// Initialize opens the SQLite database and creates the schema
func Initialize(dataSourceName string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dataSourceName+"?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.createSchema(); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// createSchema initializes the database tables
func (db *DB) createSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		description TEXT NOT NULL,
		completed INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL
	);`

	_, err := db.conn.Exec(query)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetAllTasks retrieves all tasks from the database
func (db *DB) GetAllTasks() ([]models.Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	rows, err := db.conn.Query("SELECT id, description, completed, created_at FROM tasks ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var completed int
		var createdAt string

		if err := rows.Scan(&t.ID, &t.Description, &completed, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		t.Completed = completed != 0
		t.CreatedAt, err = time.Parse("2006-01-02T15:04:05Z", createdAt)
		if err != nil {
			// Try alternative format
			t.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
			if err != nil {
				t.CreatedAt = time.Now().UTC()
			}
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	return tasks, nil
}

// CreateTask inserts a new task into the database
func (db *DB) CreateTask(description string) (*models.Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	task := &models.Task{
		ID:          uuid.New().String(),
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now().UTC(),
	}

	stmt, err := db.conn.Prepare("INSERT INTO tasks (id, description, completed, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	createdAtStr := task.CreatedAt.Format("2006-01-02T15:04:05Z")
	_, err = stmt.Exec(task.ID, task.Description, 0, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	return task, nil
}

// UpdateTask updates the completed status of a task
func (db *DB) UpdateTask(id string, completed bool) (*models.Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Use transaction for optimistic update
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check task exists
	var task models.Task
	var comp int
	var createdAt string

	err = tx.QueryRow("SELECT id, description, completed, created_at FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.Description, &comp, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	// Update completed status
	completedInt := 0
	if completed {
		completedInt = 1
	}

	stmt, err := tx.Prepare("UPDATE tasks SET completed = ? WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(completedInt, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	task.Completed = completed
	task.CreatedAt, err = time.Parse("2006-01-02T15:04:05Z", createdAt)
	if err != nil {
		task.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			task.CreatedAt = time.Now().UTC()
		}
	}

	return &task, nil
}

// DeleteTask removes a task from the database by ID
func (db *DB) DeleteTask(id string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.conn.Prepare("DELETE FROM tasks WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return false, fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}
