package services

import (
	"fmt"
	"log"
	"strings"

	"file_processor/models"
)

type ProcessorService struct {
	fileService  *FileService
	imageService *ImageService
}

func NewProcessorService(fileService *FileService, imageService *ImageService) *ProcessorService {
	return &ProcessorService{
		fileService:  fileService,
		imageService: imageService,
	}
}

func (s *ProcessorService) ProcessFile(filepath string) (*models.ProcessingResult, error) {
	log.Printf("Processing file: %s", filepath)

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filepath), ".pdf") {
		return nil, fmt.Errorf("only PDF files are supported")
	}

	// Process the PDF
	text, images, err := s.fileService.ProcessPDF(filepath)
	if err != nil {
		return nil, fmt.Errorf("error processing PDF: %v", err)
	}

	result := &models.ProcessingResult{
		FileType: "pdf",
		Text:     text,
		Images:   images,
	}

	return result, nil
}
