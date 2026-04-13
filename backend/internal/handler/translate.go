package handler

import (
	"encoding/json"
	"net/http"

	"deeplx-web/internal/models"
	"deeplx-web/internal/service"
)

type TranslateHandler struct {
	deeplx *service.DeepLXService
}

func NewTranslateHandler(deeplx *service.DeepLXService) *TranslateHandler {
	return &TranslateHandler{
		deeplx: deeplx,
	}
}

func (h *TranslateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	var req models.TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate text length
	if len(req.Text) == 0 {
		h.writeError(w, http.StatusBadRequest, "EMPTY_TEXT", "Text cannot be empty")
		return
	}
	if len(req.Text) > 10000 {
		h.writeError(w, http.StatusBadRequest, "TEXT_TOO_LONG", "Text exceeds 10000 character limit")
		return
	}

	// Set defaults
	if req.SourceLang == "" {
		req.SourceLang = "auto"
	}
	if req.TargetLang == "" {
		req.TargetLang = "ZH"
	}

	// Translate
	result, err := h.deeplx.Translate(req.Text, req.SourceLang, req.TargetLang)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "TRANSLATION_FAILED", err.Error())
		return
	}

	// Write response
	h.writeSuccess(w, result.Data, result.ID, result.Alternatives)
}

func (h *TranslateHandler) writeSuccess(w http.ResponseWriter, result string, id int, alternatives []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"result":      result,
			"id":          id,
			"alternatives": alternatives,
		},
	})
}

func (h *TranslateHandler) writeError(w http.ResponseWriter, status int, code, message string) {
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