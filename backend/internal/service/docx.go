package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"deeplx-web/pkg/utils"

	"go.uber.org/zap"
)

const (
	maxBatchSize        = 2000
	delayBetweenBatches = 2 * time.Second
	subprocessTimeout   = 5 * time.Minute
)

// DocxService handles DOCX and TXT translation
type DocxService struct {
	deeplx *DeepLXService
}

// ProgressCallback is called during document processing to report progress
type ProgressCallback func(currentBatch, totalBatches int)

func NewDocxService(deeplx *DeepLXService) *DocxService {
	return &DocxService{
		deeplx: deeplx,
	}
}

// pythonHelperPath returns the path to the docx_helper.py script
func pythonHelperPath() string {
	// Look relative to the binary location, then fall back to ./scripts/
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "scripts", "docx_helper.py")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(".", "scripts", "docx_helper.py")
}

// extractParagraphs calls the Python helper to extract all paragraph texts from a DOCX file
func extractParagraphs(docxPath string) ([]map[string]interface{}, error) {
	helper := pythonHelperPath()
	ctx, cancel := context.WithTimeout(context.Background(), subprocessTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", helper, "extract", docxPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("extract failed: %s: %w", stderr.String(), err)
	}

	var paragraphs []map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &paragraphs); err != nil {
		return nil, fmt.Errorf("failed to parse extract output: %w", err)
	}

	return paragraphs, nil
}

// replaceParagraphs calls the Python helper to write translations back to a DOCX file
func replaceParagraphs(inputPath, outputPath string, translations map[int]string) error {
	helper := pythonHelperPath()

	// Write translations to temp JSON file
	transPath := outputPath + ".trans.json"
	transFile, err := os.Create(transPath)
	if err != nil {
		return fmt.Errorf("failed to create translations file: %w", err)
	}
	defer os.Remove(transPath)

	// Convert int keys to string keys for JSON
	transMap := make(map[string]string, len(translations))
	for k, v := range translations {
		transMap[fmt.Sprintf("%d", k)] = v
	}

	if err := json.NewEncoder(transFile).Encode(transMap); err != nil {
		transFile.Close()
		return fmt.Errorf("failed to write translations: %w", err)
	}
	transFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), subprocessTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", helper, "replace", inputPath, outputPath, transPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("replace failed: %s: %w", stderr.String(), err)
	}

	return nil
}

// paragraphItem holds a paragraph's index and text for batching
type paragraphItem struct {
	index int
	text  string
}

// translateBatchWithIndex translates a batch of paragraphs using numbered markers,
// returns a map of index -> translated text
func (s *DocxService) translateBatchWithIndex(batch []paragraphItem, sourceLang, targetLang string) (map[int]string, error) {
	// Build combined text with special markers: ⟨⟨001⟩⟩ text\n⟨⟨002⟩⟩ text\n...
	// Using ⟨⟨⟩⟩ (mathematical angle brackets) to avoid collision with normal text
	var lines []string
	for _, p := range batch {
		lines = append(lines, fmt.Sprintf("⟨⟨%03d⟩⟩ %s", p.index, p.text))
	}
	combined := strings.Join(lines, "\n")

	utils.Logger.Debug("Translating batch",
		zap.Int("paragraph_count", len(batch)),
		zap.Int("combined_length", len(combined)),
	)

	result, err := s.deeplx.TranslateWithRetry(combined, sourceLang, targetLang)
	if err != nil {
		return nil, err
	}

	// Parse numbered markers from translated text
	// Pattern: ⟨⟨NNN⟩⟩ translated text
	re := regexp.MustCompile(`⟨⟨(\d{3})⟩⟩\s*`)
	matches := re.FindAllStringSubmatchIndex(result.Data, -1)

	// Cleanup: DeepL sometimes leaks ⟩ characters after markers
	cleanMarkerRemnant := regexp.MustCompile(`^\s*[⟩⟩]+\s*`)

	translations := make(map[int]string)

	if len(matches) == len(batch) {
		// Marker count matches — verify that the parsed numbers correspond to batch indices
		orderCorrect := true
		for i, m := range matches {
			numStr := result.Data[m[2]:m[3]]
			var num int
			fmt.Sscanf(numStr, "%d", &num)
			if num != batch[i].index {
				orderCorrect = false
				break
			}
		}

		if !orderCorrect {
			utils.Logger.Warn("Marker numbers changed in translation, using positional mapping",
				zap.Int("batch_size", len(batch)),
			)
		}

		// Extract translated text between markers using positional mapping
		// Always use batch[i].index as key (not the parsed num) for safety
		for i, m := range matches {
			start := m[1] // after the marker
			var end int
			if i+1 < len(matches) {
				end = matches[i+1][0] // before next marker
			} else {
				end = len(result.Data)
			}

			text := strings.TrimSpace(result.Data[start:end])
			text = cleanMarkerRemnant.ReplaceAllString(text, "")
			translations[batch[i].index] = text
		}
	} else {
		// Fallback: markers were corrupted, split by remaining markers or newlines
		utils.Logger.Warn("Batch marker count mismatch, using fallback split",
			zap.Int("expected", len(batch)),
			zap.Int("found", len(matches)),
		)

		// Split translated text by newline and match 1:1 with original batch order
		parts := strings.Split(result.Data, "\n")
		translatedParts := make([]string, 0, len(parts))
		for _, part := range parts {
			cleaned := strings.TrimSpace(re.ReplaceAllString(part, ""))
			cleaned = cleanMarkerRemnant.ReplaceAllString(cleaned, "")
			if cleaned != "" {
				translatedParts = append(translatedParts, cleaned)
			}
		}

		// Match by order
		for i, p := range batch {
			if i < len(translatedParts) {
				translations[p.index] = translatedParts[i]
			}
		}
	}

	return translations, nil
}

// ProcessDocx reads a docx file, translates all text, and returns a new docx
func (s *DocxService) ProcessDocx(r io.Reader, sourceLang, targetLang string, progressCb ProgressCallback) ([]byte, error) {
	utils.Logger.Info("Processing DOCX document (via python-docx helper)",
		zap.String("source_lang", sourceLang),
		zap.String("target_lang", targetLang),
	)

	// Save input to temp file
	inFile, err := os.CreateTemp("", "docx_input_*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	inPath := inFile.Name()
	defer os.Remove(inPath)

	if _, err := io.Copy(inFile, r); err != nil {
		inFile.Close()
		return nil, fmt.Errorf("failed to save input: %w", err)
	}
	inFile.Close()

	// Phase 1: Extract paragraphs
	paragraphs, err := extractParagraphs(inPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract paragraphs: %w", err)
	}

	// Build list of translatable paragraphs (skip empty)
	var items []paragraphItem
	for _, p := range paragraphs {
		idx, _ := p["index"].(float64)
		text, _ := p["text"].(string)
		if strings.TrimSpace(text) != "" {
			items = append(items, paragraphItem{index: int(idx), text: text})
		}
	}

	utils.Logger.Info("Extracted paragraphs",
		zap.Int("total", len(paragraphs)),
		zap.Int("translatable", len(items)),
	)

	if len(items) == 0 {
		utils.Logger.Warn("No translatable text found in DOCX")
		return nil, fmt.Errorf("no text content found in document")
	}

	// Phase 2: Split into batches by character size, translate each batch
	var batches [][]paragraphItem
	var currentBatch []paragraphItem
	currentSize := 0

	for _, item := range items {
		// Estimate marker + newline overhead (~10 chars per paragraph)
		itemSize := len(item.text) + 10
		if currentSize+itemSize > maxBatchSize && len(currentBatch) > 0 {
			batches = append(batches, currentBatch)
			currentBatch = []paragraphItem{item}
			currentSize = itemSize
		} else {
			currentBatch = append(currentBatch, item)
			currentSize += itemSize
		}
	}
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	utils.Logger.Info("Split into batches",
		zap.Int("total_batches", len(batches)),
		zap.Int("total_paragraphs", len(items)),
	)

	translations := make(map[int]string)
	for i, batch := range batches {
		utils.Logger.Info("Processing batch",
			zap.Int("batch_number", i+1),
			zap.Int("batch_size", len(batch)),
		)

		batchResult, err := s.translateBatchWithIndex(batch, sourceLang, targetLang)
		if err != nil {
			utils.Logger.Error("Failed to translate batch, falling back to per-paragraph",
				zap.Int("batch_number", i+1),
				zap.Error(err),
			)
			// Fallback: translate each paragraph in this batch individually
			for _, item := range batch {
				result, err := s.deeplx.TranslateWithRetry(item.text, sourceLang, targetLang)
				if err != nil {
					utils.Logger.Warn("Failed to translate paragraph, skipping",
						zap.Int("index", item.index),
						zap.Error(err),
					)
					continue
				}
				translations[item.index] = result.Data
			}
		} else {
			for k, v := range batchResult {
				translations[k] = v
			}
		}

		// Report progress
		if progressCb != nil {
			progressCb(i+1, len(batches))
		}

		// Delay between batches
		if i < len(batches)-1 {
			time.Sleep(delayBetweenBatches)
		}
	}

	utils.Logger.Info("Translation phase completed",
		zap.Int("translated", len(translations)),
		zap.Int("total", len(items)),
	)

	// Phase 3: Replace paragraphs
	outFile, err := os.CreateTemp("", "docx_output_*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create output temp file: %w", err)
	}
	outPath := outFile.Name()
	outFile.Close()
	defer os.Remove(outPath)

	if err := replaceParagraphs(inPath, outPath, translations); err != nil {
		return nil, fmt.Errorf("failed to replace paragraphs: %w", err)
	}

	// Read output file
	result, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	utils.Logger.Info("DOCX translation completed",
		zap.Int("output_size", len(result)),
	)

	return result, nil
}

// splitTextIntoChunks splits text into chunks of max size, trying to preserve sentence boundaries
func (s *DocxService) splitTextIntoChunks(text string, maxSize int) []string {
	var chunks []string
	currentChunk := ""
	currentSize := 0

	// Split by newlines first to preserve line structure
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		lineSize := len(line)
		// Add 1 for newline character
		if currentSize+lineSize+1 > maxSize && len(currentChunk) > 0 {
			chunks = append(chunks, strings.TrimRight(currentChunk, "\n"))
			currentChunk = line + "\n"
			currentSize = lineSize + 1
		} else {
			currentChunk += line + "\n"
			currentSize += lineSize + 1
		}
	}

	// Add the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.TrimRight(currentChunk, "\n"))
	}

	return chunks
}

// ProcessTxt reads a txt file, translates the content, and returns the translated text
func (s *DocxService) ProcessTxt(r io.Reader, sourceLang, targetLang string, progressCb ProgressCallback) ([]byte, error) {
	utils.Logger.Info("Processing TXT document",
		zap.String("source_lang", sourceLang),
		zap.String("target_lang", targetLang),
	)

	content, err := io.ReadAll(r)
	if err != nil {
		utils.Logger.Error("Failed to read TXT input", zap.Error(err))
		return nil, fmt.Errorf("failed to read txt: %w", err)
	}

	text := string(content)
	if strings.TrimSpace(text) == "" {
		utils.Logger.Warn("No text content found in TXT file")
		return nil, fmt.Errorf("no text content found in file")
	}

	utils.Logger.Debug("Translating TXT content",
		zap.Int("char_count", len(text)),
	)

	// Split into chunks if text is too long
	var chunks []string
	if len(text) > maxBatchSize {
		chunks = s.splitTextIntoChunks(text, maxBatchSize)
		utils.Logger.Info("Split text into chunks",
			zap.Int("total_chunks", len(chunks)),
			zap.Int("total_chars", len(text)),
		)
	} else {
		chunks = []string{text}
	}

	var translatedChunks []string
	for i, chunk := range chunks {
		if len(chunks) > 1 {
			utils.Logger.Info("Translating TXT chunk",
				zap.Int("chunk_number", i+1),
				zap.Int("chunk_size", len(chunk)),
			)
		}

		result, err := s.deeplx.TranslateWithRetry(chunk, sourceLang, targetLang)
		if err != nil {
			utils.Logger.Error("Failed to translate TXT content",
				zap.Int("chunk_number", i+1),
				zap.Error(err),
			)
			return nil, fmt.Errorf("chunk %d translation failed: %w", i+1, err)
		}

		translatedChunks = append(translatedChunks, result.Data)

		// Report progress
		if progressCb != nil && len(chunks) > 1 {
			progressCb(i+1, len(chunks))
		}

		// Add delay between chunks (except last)
		if i < len(chunks)-1 {
			utils.Logger.Debug("Waiting before next chunk",
				zap.Duration("delay", delayBetweenBatches),
			)
			time.Sleep(delayBetweenBatches)
		}
	}

	utils.Logger.Info("TXT translation completed successfully",
		zap.Int("result_length", len(strings.Join(translatedChunks, ""))),
	)

	return []byte(strings.Join(translatedChunks, "\n")), nil
}
