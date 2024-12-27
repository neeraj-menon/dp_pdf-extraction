package services

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("Starting image analysis with %d bytes of image data", len(imageData))

	ctx := context.Background()
	model := s.llmClient.GenerativeModel("gemini-1.5-flash")

	prompt := []genai.Part{
		genai.ImageData("image/jpeg", imageData),
		genai.Text("Extract information from the image, including any text or numerical content if present."),
	}

	log.Printf("Sending image to Gemini for analysis")
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Printf("Error generating content from image: %v", err)
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if resp.Candidates == nil {
		log.Printf("No response generated from image analysis")
		return "", fmt.Errorf("no response generated")
	}

	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		result := string(textPart)
		log.Printf("Successfully extracted information from image: %d characters", len(result))
		return result, nil
	}

	log.Printf("Unexpected response format from image analysis")
	return "", fmt.Errorf("unexpected response format")
}
