package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"deeplx-web/internal/models"
	"deeplx-web/internal/service"
	"deeplx-web/pkg/utils"

	"go.uber.org/zap"
)

// DocumentWorker handles document translation tasks
type DocumentWorker struct {
	taskManager *service.TaskManager
	docxService *service.DocxService
	uploadPath  string
	taskQueue   chan *models.Task
	workers     int
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewDocumentWorker creates a new document worker
func NewDocumentWorker(taskManager *service.TaskManager, docxService *service.DocxService, uploadPath string, workers int) *DocumentWorker {
	return &DocumentWorker{
		taskManager: taskManager,
		docxService: docxService,
		uploadPath:  uploadPath,
		taskQueue:   make(chan *models.Task, 100), // Buffer for 100 tasks
		workers:     workers,
		stopCh:      make(chan struct{}),
	}
}

// Start starts the worker pool
func (w *DocumentWorker) Start() {
	utils.Logger.Info("Starting document worker pool",
		zap.Int("workers", w.workers),
	)

	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	utils.Logger.Info("Document worker pool started", zap.Int("workers", w.workers))
}

// Stop stops the worker pool gracefully
func (w *DocumentWorker) Stop() {
	utils.Logger.Info("Stopping document worker pool")
	close(w.stopCh)
	w.wg.Wait()
	utils.Logger.Info("Document worker pool stopped")
}

// Enqueue adds a task to the processing queue
func (w *DocumentWorker) Enqueue(task *models.Task) error {
	select {
	case w.taskQueue <- task:
		utils.Logger.Info("Task enqueued",
			zap.Int64("task_id", task.ID),
			zap.String("file_name", task.FileName),
		)
		return nil
	default:
		// Queue is full
		return fmt.Errorf("task queue is full")
	}
}

// worker processes tasks from the queue
func (w *DocumentWorker) worker(id int) {
	defer w.wg.Done()

	utils.Logger.Info("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-w.stopCh:
			utils.Logger.Info("Worker stopping", zap.Int("worker_id", id))
			return
		case task := <-w.taskQueue:
			utils.Logger.Info("Worker processing task",
				zap.Int("worker_id", id),
				zap.Int64("task_id", task.ID),
			)
			w.processTask(task)
		}
	}
}

// processTask processes a single translation task
func (w *DocumentWorker) processTask(task *models.Task) {
	// Update task status to processing
	if err := w.taskManager.UpdateTaskStatus(task.ID, models.TaskStatusProcessing); err != nil {
		utils.Logger.Error("Failed to update task status",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
		return
	}

	// Construct the uploaded file path
	uploadedFilePath := filepath.Join(w.uploadPath, "temp_"+fmt.Sprint(task.ID)+"_"+task.FileName)

	// Read the uploaded file
	file, err := os.Open(uploadedFilePath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open uploaded file: %v", err)
		utils.Logger.Error("Failed to open uploaded file",
			zap.Int64("task_id", task.ID),
			zap.String("file_path", uploadedFilePath),
			zap.Error(err),
		)
		w.taskManager.FailTask(task.ID, errMsg)
		return
	}
	defer file.Close()

	// Process based on file type
	var result []byte
	var resultFileName string

	if task.FileType == ".docx" {
		resultFileName = fmt.Sprintf("%d.docx", task.ID)
		result, err = w.docxService.ProcessDocx(
			file,
			task.SourceLang,
			task.TargetLang,
			func(currentBatch, totalBatches int) {
				// Progress callback
				progress := int(float64(currentBatch) / float64(totalBatches) * 100)
				w.taskManager.UpdateTaskProgress(task.ID, progress, currentBatch, totalBatches)
			},
		)
	} else if task.FileType == ".txt" {
		resultFileName = fmt.Sprintf("%d.txt", task.ID)
		result, err = w.docxService.ProcessTxt(
			file,
			task.SourceLang,
			task.TargetLang,
			func(currentBatch, totalBatches int) {
				// Progress callback
				progress := int(float64(currentBatch) / float64(totalBatches) * 100)
				w.taskManager.UpdateTaskProgress(task.ID, progress, currentBatch, totalBatches)
			},
		)
	} else {
		errMsg := fmt.Sprintf("Unsupported file type: %s", task.FileType)
		utils.Logger.Error("Unsupported file type",
			zap.Int64("task_id", task.ID),
			zap.String("file_type", task.FileType),
		)
		w.taskManager.FailTask(task.ID, errMsg)
		return
	}

	if err != nil {
		errMsg := fmt.Sprintf("Translation failed: %v", err)
		utils.Logger.Error("Translation failed",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
		w.taskManager.FailTask(task.ID, errMsg)
		return
	}

	// Save result to file
	resultFilePath := filepath.Join(w.uploadPath, resultFileName)
	if err := os.WriteFile(resultFilePath, result, 0644); err != nil {
		errMsg := fmt.Sprintf("Failed to save result file: %v", err)
		utils.Logger.Error("Failed to save result file",
			zap.Int64("task_id", task.ID),
			zap.String("file_path", resultFilePath),
			zap.Error(err),
		)
		w.taskManager.FailTask(task.ID, errMsg)
		return
	}

	// Clean up uploaded file
	if err := os.Remove(uploadedFilePath); err != nil {
		utils.Logger.Warn("Failed to remove uploaded file",
			zap.Int64("task_id", task.ID),
			zap.String("file_path", uploadedFilePath),
			zap.Error(err),
		)
	}

	// Mark task as completed
	if err := w.taskManager.CompleteTask(task.ID, resultFilePath); err != nil {
		utils.Logger.Error("Failed to mark task as completed",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
		return
	}

	utils.Logger.Info("Task processed successfully",
		zap.Int64("task_id", task.ID),
		zap.String("result_file_path", resultFilePath),
	)
}

// GetQueueSize returns the current queue size
func (w *DocumentWorker) GetQueueSize() int {
	return len(w.taskQueue)
}
