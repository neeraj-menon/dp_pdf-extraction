package services

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

type ImageService struct {
	llmClient *genai.Client
}

func NewImageService(llmClient *genai.Client) *ImageService {
	return &ImageService{
		llmClient: llmClient,
	}
}

func (s *ImageService) ExtractInformationFromImage(imageData []byte) (string, error) {
	ctx := context.Background()
	model := s.llmClient.GenerativeModel("gemini-1.5-flash")

	prompt := []genai.Part{
		genai.ImageData("image/jpeg", imageData),
		genai.Text("Please describe what you see in this image in detail, including any text content if present."),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	// Get the text from the response
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(textPart), nil
	}

	return "", fmt.Errorf("unexpected response format")
}
