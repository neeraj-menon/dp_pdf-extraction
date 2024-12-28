package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type FileService struct {
	imageService  *ImageService
	geminiService *GeminiService
}

func NewFileService(imageService *ImageService, geminiService *GeminiService) *FileService {
	return &FileService{
		imageService:  imageService,
		geminiService: geminiService,
	}
}

func (s *FileService) ProcessPDF(filePath string) (string, []string, error) {
	log.Printf("Processing PDF file: %s", filePath)

	// Create temp directory for images
	tempDir := filepath.Join("temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", nil, fmt.Errorf("error creating temp directory: %v", err)
	}

	// Convert PDF to images using Python script
	cmd := exec.Command("python", "scripts/pdf_to_image.py", filePath, tempDir)
	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("error converting PDF to images: %v", err)
	}

	// Get list of generated images
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return "", nil, fmt.Errorf("error reading temp directory: %v", err)
	}

	var imagePaths []string
	for _, file := range files {
		if !file.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(tempDir, file.Name()))
		}
	}

	// Process images with Gemini Vision
	var fullText string
	for _, imagePath := range imagePaths {
		text, err := s.imageService.ProcessImage(imagePath)
		if err != nil {
			return "", nil, fmt.Errorf("error processing image %s: %v", imagePath, err)
		}
		fullText += text + "\n"
	}

	// Process the extracted text with Gemini
	if s.geminiService != nil {
		if _, err := s.geminiService.ProcessDocument(context.Background(), fullText, filepath.Base(filePath)); err != nil {
			log.Printf("Error processing document with Gemini: %v", err)
			// Continue with raw text even if Gemini processing fails
		} else {
			log.Printf("Successfully processed document with Gemini")
		}
	}

	return fullText, imagePaths, nil
}
