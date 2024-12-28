package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"file_processor/models"
)

type GeminiService struct {
	client      *genai.Client
	model       *genai.GenerativeModel
	ruleService *RuleService
}

func NewGeminiService(apiKey string, ruleService *RuleService) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(0.2)
	return &GeminiService{
		client:      client,
		model:       model,
		ruleService: ruleService,
	}, nil
}

// GetClient returns the Gemini client for use by other services
func (s *GeminiService) GetClient() *genai.Client {
	return s.client
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(b)
}

func (s *GeminiService) ProcessDocument(ctx context.Context, text, filename string) (*models.StructuredResponse, error) {
	log.Printf("ProcessDocument called with filename: %s (text length: %d)", filename, len(text))

	prompt := fmt.Sprintf(`Analyze the following document text and return ONLY a JSON object with this exact structure (no other text):

{
    "document_id": "",
    "filename": "",
    "is_ruleset": false,
    "file_type": "",
    "upload_date": "",
    "content": {
        "extracted_content": {}
    }
}

Guidelines:
1. Keep the structure exactly as shown
2. Leave document_id, filename, and upload_date as empty strings
3. Set is_ruleset to true only if the document contains rules or policies
4. For file_type, use one of: Medical Report, Legal Document, Financial Statement, Policy Document, Technical Manual, General Text
5. In extracted_content, include relevant key-value pairs from the document

Text to analyze:
%s`, text)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Get the response text and ensure it's clean JSON
	textContent, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response format from Gemini")
	}

	// Clean the response text to ensure it's valid JSON
	jsonStr := string(textContent)
	jsonStr = strings.TrimSpace(jsonStr)

	// Remove any potential markdown code block markers
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	log.Printf("Cleaned JSON response: %s", jsonStr)

	// Parse the JSON response
	var structuredResp models.StructuredResponse
	if err := json.Unmarshal([]byte(jsonStr), &structuredResp); err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		return nil, fmt.Errorf("failed to parse JSON: %v\nResponse was: %s", err, jsonStr)
	}

	log.Printf("Before metadata - DocumentID: %s, Filename: %s, UploadDate: %s",
		structuredResp.DocumentID, structuredResp.Filename, structuredResp.UploadDate)

	// Generate new UUID
	newUUID := generateUUID()
	log.Printf("Generated UUID: %s", newUUID)

	// Add metadata
	structuredResp.DocumentID = newUUID
	structuredResp.Filename = filename
	structuredResp.UploadDate = time.Now().Format(time.RFC3339)

	log.Printf("After metadata - DocumentID: %s, Filename: %s, UploadDate: %s",
		structuredResp.DocumentID, structuredResp.Filename, structuredResp.UploadDate)

	// Log the complete structured response
	respJSON, _ := json.MarshalIndent(structuredResp, "", "  ")
	log.Printf("Final structured response:\n%s", string(respJSON))

	// Determine whether to add as rule or data based on is_ruleset flag
	if s.ruleService != nil {
		var err error
		if structuredResp.IsRuleset {
			err = s.ruleService.AddRule(ctx, &structuredResp)
		} else {
			err = s.ruleService.AddData(ctx, &structuredResp)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to process document: %v", err)
		}
	}

	return &structuredResp, nil
}

func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
