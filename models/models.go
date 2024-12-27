package models

type ProcessingResult struct {
	FileType      string            `json:"file_type"`
	Text          string            `json:"text"`
	Images        []string          `json:"images,omitempty"`
	ImageAnalysis map[string]string `json:"image_analysis,omitempty"`
}

type URLInput struct {
	URL string `json:"url" binding:"required"`
}

type ContentStruct struct {
	ExtractedContent map[string]interface{} `json:"extracted_content"`
}

type StructuredResponse struct {
	DocumentID  string        `json:"document_id"`
	Filename    string        `json:"filename"`
	IsRuleset   bool          `json:"is_ruleset"`
	FileType    string        `json:"file_type"`
	UploadDate  string        `json:"upload_date"`
	Content     ContentStruct `json:"content"`
}
