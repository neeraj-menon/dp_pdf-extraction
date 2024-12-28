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

type PersonalInformation struct {
	Name     string `json:"name,omitempty"`
	Location string `json:"location,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
}

type TechnicalSkills struct {
	Languages    []string `json:"languages,omitempty"`
	Technologies []string `json:"technologies,omitempty"`
}

type Education struct {
	Degree     string `json:"degree,omitempty"`
	University string `json:"university,omitempty"`
	Location   string `json:"location,omitempty"`
	Years      string `json:"years,omitempty"`
	GPA        string `json:"gpa,omitempty"`
}

type WorkExperience struct {
	Company          string   `json:"company,omitempty"`
	Title           string   `json:"title,omitempty"`
	Years           string   `json:"years,omitempty"`
	Responsibilities []string `json:"responsibilities,omitempty"`
}

type Project struct {
	Name        string `json:"name,omitempty"`
	Year        string `json:"year,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type AdditionalDetails struct {
	Certifications []string `json:"certifications,omitempty"`
	Awards         []string `json:"awards,omitempty"`
	Languages      []string `json:"languages,omitempty"`
}

type ExtractedContent struct {
	Columns              []string            `json:"columns,omitempty"`
	PersonalInformation  *PersonalInformation `json:"personal_information,omitempty"`
	TechnicalSkills     *TechnicalSkills    `json:"technical_skills,omitempty"`
	Education           []Education         `json:"education,omitempty"`
	WorkExperience      []WorkExperience    `json:"work_experience,omitempty"`
	Projects            []Project           `json:"projects,omitempty"`
	AdditionalDetails   *AdditionalDetails  `json:"additional_details,omitempty"`
}

type ContentStruct struct {
	ExtractedContent ExtractedContent `json:"extracted_content"`
}

type StructuredResponse struct {
	DocumentID  string        `json:"document_id"`
	Filename    string        `json:"filename"`
	IsRuleset   bool          `json:"is_ruleset"`
	FileType    string        `json:"file_type"`
	UploadDate  string        `json:"upload_date"`
	Content     ContentStruct `json:"content"`
}
