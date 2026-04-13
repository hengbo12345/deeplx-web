package service

import (
	"fmt"
	"sync"
	"time"

	"deeplx-web/internal/models"
	"deeplx-web/pkg/utils"

	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
)

// TaskManager manages translation tasks in memory
type TaskManager struct {
	tasks         sync.Map // Key: int64 (task ID), Value: *models.Task
	snowflakeNode *snowflake.Node
}

// NewTaskManager creates a new task manager
func NewTaskManager() (*TaskManager, error) {
	// Initialize Snowflake node with node ID 1
	// In a distributed system, each instance should have a unique node ID
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize snowflake node: %w", err)
	}

	tm := &TaskManager{
		snowflakeNode: node,
	}

	// Start periodic cleanup of expired tasks
	go tm.periodicCleanup()

	utils.Logger.Info("Task manager initialized")
	return tm, nil
}

// CreateTask creates a new translation task
func (tm *TaskManager) CreateTask(fileName, fileType, sourceLang, targetLang string) (*models.Task, error) {
	// Generate unique task ID using Snowflake
	taskID := tm.snowflakeNode.Generate().Int64()

	now := time.Now()
	task := &models.Task{
		ID:         taskID,
		Status:     models.TaskStatusPending,
		Progress:   0,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		FileName:   fileName,
		FileType:   fileType,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Store task
	tm.tasks.Store(taskID, task)

	utils.Logger.Info("Task created",
		zap.Int64("task_id", taskID),
		zap.String("file_name", fileName),
		zap.String("file_type", fileType),
		zap.String("source_lang", sourceLang),
		zap.String("target_lang", targetLang),
	)

	return task, nil
}

// GetTask retrieves a task by ID
func (tm *TaskManager) GetTask(taskID int64) (*models.Task, error) {
	value, ok := tm.tasks.Load(taskID)
	if !ok {
		return nil, fmt.Errorf("task not found: %d", taskID)
	}

	task, ok := value.(*models.Task)
	if !ok {
		return nil, fmt.Errorf("invalid task type for ID: %d", taskID)
	}

	return task, nil
}

// UpdateTaskStatus updates the status of a task
func (tm *TaskManager) UpdateTaskStatus(taskID int64, status models.TaskStatus) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	if status == models.TaskStatusCompleted {
		now := time.Now()
		task.CompletedAt = &now
		task.Progress = 100
	}

	utils.Logger.Debug("Task status updated",
		zap.Int64("task_id", taskID),
		zap.String("status", string(status)),
	)

	return nil
}

// UpdateTaskProgress updates the progress of a task
func (tm *TaskManager) UpdateTaskProgress(taskID int64, progress, currentBatch, totalBatches int) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Progress = progress
	task.CurrentBatch = currentBatch
	task.TotalBatches = totalBatches
	task.UpdatedAt = time.Now()

	// Automatically update status to processing if still pending
	if task.Status == models.TaskStatusPending {
		task.Status = models.TaskStatusProcessing
	}

	utils.Logger.Debug("Task progress updated",
		zap.Int64("task_id", taskID),
		zap.Int("progress", progress),
		zap.Int("current_batch", currentBatch),
		zap.Int("total_batches", totalBatches),
	)

	return nil
}

// CompleteTask marks a task as completed with the result file path
func (tm *TaskManager) CompleteTask(taskID int64, resultFilePath string) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	now := time.Now()
	task.Status = models.TaskStatusCompleted
	task.Progress = 100
	task.ResultFilePath = resultFilePath
	task.CompletedAt = &now
	task.UpdatedAt = now

	utils.Logger.Info("Task completed",
		zap.Int64("task_id", taskID),
		zap.String("result_file_path", resultFilePath),
		zap.Duration("duration", now.Sub(task.CreatedAt)),
	)

	return nil
}

// FailTask marks a task as failed with an error message
func (tm *TaskManager) FailTask(taskID int64, errMsg string) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = models.TaskStatusFailed
	task.Error = errMsg
	task.UpdatedAt = time.Now()

	utils.Logger.Error("Task failed",
		zap.Int64("task_id", taskID),
		zap.String("error", errMsg),
	)

	return nil
}

// DeleteTask removes a task from memory
func (tm *TaskManager) DeleteTask(taskID int64) {
	tm.tasks.Delete(taskID)
	utils.Logger.Debug("Task deleted", zap.Int64("task_id", taskID))
}

// CleanupExpiredTasks removes tasks older than maxAge
func (tm *TaskManager) CleanupExpiredTasks(maxAge time.Duration) {
	now := time.Now()
	expiredTasks := make([]int64, 0)

	// Find expired tasks
	tm.tasks.Range(func(key, value interface{}) bool {
		taskID := key.(int64)
		task := value.(*models.Task)

		// Only cleanup completed or failed tasks older than maxAge
		if task.IsFinalStatus() && now.Sub(task.UpdatedAt) > maxAge {
			expiredTasks = append(expiredTasks, taskID)
		}
		return true
	})

	// Delete expired tasks
	for _, taskID := range expiredTasks {
		tm.DeleteTask(taskID)
	}

	if len(expiredTasks) > 0 {
		utils.Logger.Info("Cleaned up expired tasks",
			zap.Int("count", len(expiredTasks)),
			zap.Duration("max_age", maxAge),
		)
	}
}

// periodicCleanup runs periodic cleanup of expired tasks
func (tm *TaskManager) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		// Cleanup tasks older than 1 hour
		tm.CleanupExpiredTasks(1 * time.Hour)
	}
}

// GetTaskCount returns the current number of tasks in memory
func (tm *TaskManager) GetTaskCount() int {
	count := 0
	tm.tasks.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
