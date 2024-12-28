package services

import (
	"context"
	"fmt"
	"log"
	"os"

	genai "github.com/google/generative-ai-go/genai"
)

type ImageService struct {
	client *genai.Client
}

func NewImageService(client *genai.Client) *ImageService {
	return &ImageService{
		client: client,
	}
}

func (s *ImageService) ProcessImage(imagePath string) (string, error) {
	log.Printf("Processing image: %s", imagePath)

	// Read image file
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("error reading image file: %v", err)
	}

	// Create model for image processing
	model := s.client.GenerativeModel("gemini-1.5-flash")

	// Process image with Gemini Vision
	prompt := []genai.Part{
		genai.ImageData("image/png", imgData),
		genai.Text("Extract all text from this image, preserving formatting and structure. Include any tables, lists, or special formatting."),
	}

	resp, err := model.GenerateContent(context.Background(), prompt...)
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no text extracted from image")
	}

	// Get text content
	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	log.Printf("Successfully extracted text from image: %s", imagePath)
	return string(text), nil
}
