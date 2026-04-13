package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"deeplx-web/pkg/utils"

	"go.uber.org/zap"
)

type DeepLXService struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewDeepLXService(baseURL, token string) *DeepLXService {
	return &DeepLXService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
		token:   token,
	}
}

type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type TranslateResponse struct {
	Code         int      `json:"code"` // 200 is success, other values indicate errors
	ID           int      `json:"id"`
	Data         string   `json:"data"`                   // Translated text
	Alternatives []string `json:"alternatives,omitempty"` // Optional alternative translations
}

// TranslationResult holds the full translation result from DeepLX API
type TranslationResult struct {
	Data         string
	ID           int
	Alternatives []string
}

func (s *DeepLXService) Translate(text, sourceLang, targetLang string) (*TranslationResult, error) {
	utils.Logger.Debug("Translation request",
		zap.String("source_lang", sourceLang),
		zap.String("target_lang", targetLang),
		zap.Int("text_length", len(text)),
	)

	reqBody := TranslateRequest{
		Text:       text,
		SourceLang: sourceLang,
		TargetLang: targetLang,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		utils.Logger.Error("Failed to marshal translation request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewReader(jsonData))
	if err != nil {
		utils.Logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		utils.Logger.Error("Failed to send translation request", zap.Error(err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		utils.Logger.Error("DeepLX API returned non-OK status",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("DeepLX returned status %d: %s", resp.StatusCode, string(body))
	}

	var result TranslateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		body, _ := io.ReadAll(resp.Body)
		utils.Logger.Error("Failed to decode DeepLX response", zap.Error(err), zap.String("response", string(body)))
		return nil, fmt.Errorf("failed to decode response: %w:%s", err, string(body))
	}

	// Check DeepLX API response status code
	if result.Code != 200 {
		utils.Logger.Error("DeepLX API returned error code",
			zap.Int("code", result.Code),
			zap.String("data", result.Data),
		)
		return nil, fmt.Errorf("DeepLX API returned error code %d", result.Code)
	}

	utils.Logger.Debug("Translation successful",
		zap.Int("result_length", len(result.Data)),
		zap.Int("id", result.ID),
		zap.Int("alternatives_count", len(result.Alternatives)),
	)

	return &TranslationResult{
		Data:         result.Data,
		ID:           result.ID,
		Alternatives: result.Alternatives,
	}, nil
}

// TranslateWithRetry translates text with automatic retry on rate limit errors
func (s *DeepLXService) TranslateWithRetry(text, sourceLang, targetLang string) (*TranslationResult, error) {
	maxRetries := 3
	initialBackoff := 1 * time.Second
	maxBackoff := 10 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := s.Translate(text, sourceLang, targetLang)
		if err == nil {
			if attempt > 0 {
				utils.Logger.Info("Translation succeeded after retry",
					zap.Int("attempt", attempt+1),
				)
			}
			return result, nil
		}

		// Check if error is 429 rate limit
		if strings.Contains(err.Error(), "429") && attempt < maxRetries-1 {
			backoff := initialBackoff * time.Duration(1<<uint(attempt))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			utils.Logger.Warn("Rate limit hit, retrying",
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries exceeded")
}
