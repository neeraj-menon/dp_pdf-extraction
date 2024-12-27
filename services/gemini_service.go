package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"file_processor/models"
)

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	return &GeminiService{
		client: client,
		model:  model,
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
	prompt := fmt.Sprintf(`You are a document analysis system. Analyze the following document text and return a JSON object ONLY (no other text) with this exact structure:
{
  "document_id": "",  // leave empty, will be filled later
  "filename": "",     // leave empty, will be filled later
  "is_ruleset": true/false,  // true if document contains rules or policies
  "file_type": "",    // e.g., "Sales Report", "Medical Report", "Invoice", "Financial Statement", "Data Sheet"
  "upload_date": "",  // leave empty, will be filled later
  "content": {
    "text": "",       // the full document text
    "key_values": {}, // extracted key-value pairs from the document
    "tables": [       // array of identified tables
      {
        "table_description": "", // brief description of what the table contains
        "data": [               // 2D array representing table rows and columns
          ["Column1", "Column2", ...],  // header row
          ["Value1", "Value2", ...],    // data rows
          ...
        ]
      }
    ]
  }
}

Important:
1. For spreadsheets (Excel, CSV), focus on identifying table structures and their meanings
2. Extract any metadata or key-value pairs from headers or summary sections
3. Determine if the content represents rules/policies or data
4. Preserve table structure and relationships

Document to analyze:
%s

Remember: Return ONLY valid JSON, no other text or explanation.`, text)

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

	// Replace tabs with spaces to avoid JSON parsing issues
	jsonStr = strings.ReplaceAll(jsonStr, "\t", " ")

	// Use a decoder to handle the raw string
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber() // Handle numbers more reliably

	// Parse the JSON response
	var structuredResp models.StructuredResponse
	if err := decoder.Decode(&structuredResp); err != nil {
		// If parsing fails, try to decode into a map first
		decoder = json.NewDecoder(strings.NewReader(jsonStr))
		var rawMap map[string]interface{}
		if err := decoder.Decode(&rawMap); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %v\nResponse was: %s", err, jsonStr)
		}

		// Re-encode with proper escaping
		cleanJSON, err := json.Marshal(rawMap)
		if err != nil {
			return nil, fmt.Errorf("failed to re-encode JSON: %v", err)
		}

		// Try parsing the cleaned JSON
		if err := json.Unmarshal(cleanJSON, &structuredResp); err != nil {
			return nil, fmt.Errorf("failed to parse cleaned JSON: %v\nResponse was: %s", err, string(cleanJSON))
		}
	}

	// Add metadata
	structuredResp.DocumentID = generateUUID()
	structuredResp.Filename = filename
	structuredResp.UploadDate = time.Now().Format(time.RFC3339)

	return &structuredResp, nil
}

func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
