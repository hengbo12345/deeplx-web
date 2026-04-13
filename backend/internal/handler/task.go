package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"deeplx-web/internal/models"
	"deeplx-web/internal/service"
	"deeplx-web/internal/worker"
	"deeplx-web/pkg/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	taskManager *service.TaskManager
	documentWorker *worker.DocumentWorker
	uploadPath  string
	maxFileSize int64
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskManager *service.TaskManager, documentWorker *worker.DocumentWorker, uploadPath string, maxFileSize int64) *TaskHandler {
	return &TaskHandler{
		taskManager:     taskManager,
		documentWorker:  documentWorker,
		uploadPath:      uploadPath,
		maxFileSize:     maxFileSize,
	}
}

// CreateDocumentTask creates a new document translation task
func (h *TaskHandler) CreateDocumentTask(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 32MB to account for multipart overhead)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_FORM", "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "NO_FILE", "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > h.maxFileSize {
		h.writeError(w, http.StatusBadRequest, "FILE_TOO_LARGE",
			fmt.Sprintf("File size exceeds %d bytes limit", h.maxFileSize))
		return
	}

	// Get language parameters
	sourceLang := r.FormValue("source_lang")
	targetLang := r.FormValue("target_lang")
	if sourceLang == "" {
		sourceLang = "auto"
	}
	if targetLang == "" {
		targetLang = "ZH"
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".docx" && ext != ".txt" {
		h.writeError(w, http.StatusBadRequest, "INVALID_FILE_TYPE",
			"Unsupported file type. Only .docx and .txt are supported")
		return
	}

	// Create task
	task, err := h.taskManager.CreateTask(header.Filename, ext, sourceLang, targetLang)
	if err != nil {
		utils.Logger.Error("Failed to create task", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "TASK_CREATION_FAILED", "Failed to create translation task")
		return
	}

	// Save uploaded file with task ID prefix
	uploadedFilePath := filepath.Join(h.uploadPath, "temp_"+fmt.Sprint(task.ID)+"_"+header.Filename)
	dst, err := saveUploadedFile(file, uploadedFilePath)
	if err != nil {
		utils.Logger.Error("Failed to save uploaded file",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
		h.taskManager.FailTask(task.ID, fmt.Sprintf("Failed to save uploaded file: %v", err))
		h.writeError(w, http.StatusInternalServerError, "FILE_SAVE_FAILED", "Failed to save uploaded file")
		return
	}
	dst.Close()

	// Enqueue task for processing
	if err := h.documentWorker.Enqueue(task); err != nil {
		utils.Logger.Error("Failed to enqueue task",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
		h.taskManager.FailTask(task.ID, fmt.Sprintf("Failed to enqueue task: %v", err))
		h.writeError(w, http.StatusServiceUnavailable, "QUEUE_FULL", "Task queue is full, please try again later")
		return
	}

	// Return task ID immediately
	h.writeSuccess(w, http.StatusAccepted, map[string]interface{}{
		"task_id": fmt.Sprint(task.ID),
		"status":  string(task.Status),
		"message": "Task created successfully",
	})

	utils.Logger.Info("Document translation task created",
		zap.Int64("task_id", task.ID),
		zap.String("file_name", header.Filename),
		zap.Int64("file_size", header.Size),
	)
}

// GetTaskStatus returns the status of a translation task
func (h *TaskHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from URL parameter
	taskIDStr := chi.URLParam(r, "id")

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID format")
		return
	}

	// Get task from manager
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found or expired")
		return
	}

	// Return task status
	h.writeSuccess(w, http.StatusOK, task)
}

// DownloadTaskResult downloads the translated document
func (h *TaskHandler) DownloadTaskResult(w http.ResponseWriter, r *http.Request) {
	// Extract task ID from URL parameter
	taskIDStr := chi.URLParam(r, "id")

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_TASK_ID", "Invalid task ID format")
		return
	}

	// Get task from manager
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found or expired")
		return
	}

	// Check if task is completed
	if task.Status != models.TaskStatusCompleted {
		if task.Status == models.TaskStatusFailed {
			h.writeError(w, http.StatusBadRequest, "TASK_FAILED", "Task failed: "+task.Error)
		} else {
			h.writeError(w, http.StatusAccepted, "TASK_NOT_READY", "Task is still processing")
		}
		return
	}

	// Check if result file exists
	if task.ResultFilePath == "" {
		h.writeError(w, http.StatusInternalServerError, "NO_RESULT_FILE", "Result file not found")
		return
	}

	// Determine content type
	var contentType string
	switch task.FileType {
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		contentType = "text/plain"
	default:
		contentType = "application/octet-stream"
	}

	// Set headers for file download
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="translated_%s"`, task.FileName))

	// Serve file
	http.ServeFile(w, r, task.ResultFilePath)

	utils.Logger.Info("Task result downloaded",
		zap.Int64("task_id", task.ID),
		zap.String("file_path", task.ResultFilePath),
	)
}

// Helper methods

func (h *TaskHandler) writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *TaskHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// saveUploadedFile saves an uploaded file to the specified path
func saveUploadedFile(file io.Reader, path string) (*os.File, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// Create destination file
	dst, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(path)
		return nil, err
	}

	return dst, nil
}
