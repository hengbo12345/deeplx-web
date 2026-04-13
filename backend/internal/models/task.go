package models

import "time"

// TaskStatus represents the current status of a translation task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusExpired    TaskStatus = "expired"
)

// Task represents a document translation task
type Task struct {
	ID             int64        `json:"id"`                       // Snowflake ID
	Status         TaskStatus   `json:"status"`                   // Task status
	Progress       int          `json:"progress"`                 // Progress 0-100
	SourceLang     string       `json:"source_lang"`              // Source language
	TargetLang     string       `json:"target_lang"`              // Target language
	FileName       string       `json:"file_name"`                // Original filename
	FileType       string       `json:"file_type"`                // File type (docx, txt)
	ResultFilePath string       `json:"result_file_path,omitempty"` // Result file path
	Error          string       `json:"error,omitempty"`          // Error message if failed
	CreatedAt      time.Time    `json:"created_at"`               // Creation time
	UpdatedAt      time.Time    `json:"updated_at"`               // Last update time
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`  // Completion time

	// Progress details
	CurrentBatch int `json:"current_batch,omitempty"` // Current batch number
	TotalBatches int `json:"total_batches,omitempty"` // Total number of batches
}

// IsFinalStatus returns true if the task is in a final state (completed, failed, or expired)
func (t *Task) IsFinalStatus() bool {
	return t.Status == TaskStatusCompleted ||
		t.Status == TaskStatusFailed ||
		t.Status == TaskStatusExpired
}

// CanDownload returns true if the task result can be downloaded
func (t *Task) CanDownload() bool {
	return t.Status == TaskStatusCompleted && t.ResultFilePath != ""
}
