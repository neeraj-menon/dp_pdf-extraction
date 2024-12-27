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

func (s *ProcessorService) ProcessFile(filepath string, ext string) (*models.ProcessingResult, error) {
	log.Printf("Processing file: %s with extension: %s", filepath, ext)
	
	result := &models.ProcessingResult{
		FileType: s.determineFileType(ext),
	}

	log.Printf("Determined file type: %s", result.FileType)
	var err error

	switch strings.ToLower(ext) {
	case ".pdf":
		log.Printf("Starting PDF extraction")
		result.Text, result.Images, err = s.fileService.ExtractTextFromPDF(filepath)
		if err != nil {
			log.Printf("Error processing PDF: %v", err)
			return nil, fmt.Errorf("error processing PDF: %v", err)
		}
		log.Printf("PDF extraction complete. Text length: %d, Images found: %d", len(result.Text), len(result.Images))

	case ".doc", ".docx":
		log.Printf("Starting DOCX extraction")
		result.Text, result.Images, err = s.fileService.ExtractTextFromDOCX(filepath)
		if err != nil {
			log.Printf("Error processing DOCX: %v", err)
			return nil, fmt.Errorf("error processing DOCX: %v", err)
		}
		log.Printf("DOCX extraction complete. Text length: %d, Images found: %d", len(result.Text), len(result.Images))

	case ".xlsx", ".xls", ".csv":
		log.Printf("Starting spreadsheet extraction")
		result.Text, result.Images, err = s.fileService.ExtractTextFromFile(filepath)
		if err != nil {
			log.Printf("Error processing spreadsheet: %v", err)
			return nil, fmt.Errorf("error processing spreadsheet: %v", err)
		}
		log.Printf("Spreadsheet extraction complete. Text length: %d", len(result.Text))

	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		log.Printf("Starting image analysis")
		imgData, err := s.fileService.ReadImage(filepath)
		if err != nil {
			log.Printf("Error reading image: %v", err)
			return nil, fmt.Errorf("error reading image: %v", err)
		}
		result.Text, err = s.imageService.ExtractInformationFromImage(imgData)
		if err != nil {
			log.Printf("Error analyzing image: %v", err)
			return nil, fmt.Errorf("error analyzing image: %v", err)
		}
		log.Printf("Image analysis complete")

	default:
		log.Printf("Unsupported file type: %s", ext)
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Process any images if present
	if len(result.Images) > 0 {
		log.Printf("Processing %d images", len(result.Images))
		result.ImageAnalysis = make(map[string]string)
		for _, img := range result.Images {
			imgData, err := s.fileService.ReadImage(img)
			if err != nil {
				log.Printf("Error reading image %s: %v", img, err)
				continue
			}

			analysis, err := s.imageService.ExtractInformationFromImage(imgData)
			if err != nil {
				log.Printf("Error analyzing image %s: %v", img, err)
				continue
			}

			result.ImageAnalysis[img] = analysis
			log.Printf("Successfully analyzed image %s", img)
		}
	}

	return result, nil
}

func (s *ProcessorService) ProcessURL(url string) (*models.ProcessingResult, error) {
	// Extract text from URL
	text, images, err := s.fileService.ExtractTextFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("error processing URL: %v", err)
	}

	result := &models.ProcessingResult{
		Text:   text,
		Images: images,
	}

	return result, nil
}

func (s *ProcessorService) determineFileType(ext string) string {
	switch strings.ToLower(ext) {
	case ".pdf":
		return "PDF Document"
	case ".doc", ".docx":
		return "Word Document"
	case ".xlsx", ".xls":
		return "Excel Spreadsheet"
	case ".csv":
		return "CSV Spreadsheet"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return "Image"
	default:
		return "Unknown"
	}
}
