package models

// ProcessingResult represents the result of processing a file
type ProcessingResult struct {
	FileType string   `json:"file_type"`
	Text     string   `json:"text"`
	Images   []string `json:"images,omitempty"`
}

// StructuredResponse represents structured data extracted from a document
type StructuredResponse struct {
	DocumentID   string                 `json:"document_id"`
	Filename     string                 `json:"filename"`
	IsRuleset    bool                   `json:"is_ruleset"`
	FileType     string                 `json:"file_type"`
	UploadDate   string                 `json:"upload_date"`
	Content      Content                `json:"content"`
	RawText      string                 `json:"raw_text,omitempty"`
}

// Content represents the structured content of a document
type Content struct {
	ExtractedContent map[string]interface{} `json:"extracted_content"`
}
