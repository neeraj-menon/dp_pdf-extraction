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

type TableData struct {
	TableDescription string     `json:"table_description"`
	Data            [][]string `json:"data"`
}

type ContentStruct struct {
	Text      string            `json:"text"`
	KeyValues map[string]string `json:"key_values"`
	Tables    []TableData       `json:"tables"`
}

type StructuredResponse struct {
	DocumentID  string        `json:"document_id"`
	Filename    string        `json:"filename"`
	IsRuleset   bool         `json:"is_ruleset"`
	FileType    string       `json:"file_type"`
	UploadDate  string       `json:"upload_date"`
	Content     ContentStruct `json:"content"`
}
