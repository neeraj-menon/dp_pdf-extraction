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
	return &GeminiService{
		client:      client,
		model:       model,
		ruleService: ruleService,
	}, nil
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

	prompt := fmt.Sprintf(`You are a document analysis system. Analyze the following document text and return a JSON object ONLY (no other text) with this exact structure:
{
  "document_id": "",
  "filename": "",
  "is_ruleset": false,
  "file_type": "",
  "upload_date": "",
  "content": {
    "extracted_content": {
      key1: value1,
	  key2: value2,
	  key3: value3
    }
  }
}

Important:
1. For spreadsheets (Excel, CSV), focus on identifying table structures and their meanings
2. Extract any metadata or key-value pairs from headers or summary sections
3. Determine if the content represents rules/policies or data
4. Preserve table structure and relationships
5. Make sure you include as much information as possible from the document.

Document to analyze:
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

	log.Printf("Received JSON response: %s", jsonStr)

	// Parse the JSON response
	var structuredResp models.StructuredResponse
	if err := json.Unmarshal([]byte(jsonStr), &structuredResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v\nResponse was: %s", err, jsonStr)
	}

	// Add metadata
	structuredResp.DocumentID = generateUUID()
	structuredResp.Filename = filename
	structuredResp.UploadDate = time.Now().Format(time.RFC3339)

	// Log the details of the structured response
	log.Printf("Structured Response - IsRuleset: %v, FileType: %s",
		structuredResp.IsRuleset,
		structuredResp.FileType,
	)

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
