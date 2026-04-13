package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"deeplx-web/internal/config"
	"deeplx-web/internal/service"
)

type DocumentHandler struct {
	docx  *service.DocxService
	cfg   *config.Config
}

func NewDocumentHandler(docx *service.DocxService, cfg *config.Config) *DocumentHandler {
	return &DocumentHandler{
		docx: docx,
		cfg:  cfg,
	}
}

func (h *DocumentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32MB to account for multipart overhead)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > h.cfg.MaxFileSize {
		http.Error(w, fmt.Sprintf("File size exceeds %d bytes limit", h.cfg.MaxFileSize), http.StatusBadRequest)
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
	var result []byte
	var filename string

	switch ext {
	case ".docx":
		result, err = h.docx.ProcessDocx(file, sourceLang, targetLang, nil)
		filename = "translated_" + header.Filename
	case ".txt":
		result, err = h.docx.ProcessTxt(file, sourceLang, targetLang, nil)
		filename = "translated_" + header.Filename
	default:
		http.Error(w, "Unsupported file type. Only .docx and .txt are supported", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Translation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result)))

	w.Write(result)
}