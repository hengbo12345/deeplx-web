package handler

import (
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	token string
}

func NewAuthHandler(token string) *AuthHandler {
	return &AuthHandler{token: token}
}

type verifyRequest struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if req.Token == h.token {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]bool{
				"valid": true,
			},
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INVALID_TOKEN",
				"message": "Invalid token",
			},
		})
	}
}
