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

	var text string
	var images []string
	var err error

	// Process based on file type
	switch strings.ToLower(ext) {
	case ".pdf":
		text, images, err = s.fileService.ExtractTextFromPDF(filepath)
		if err != nil {
			if strings.Contains(err.Error(), "no text content found") {
				log.Printf("No text found in PDF, attempting image processing")
				text, images, err = s.processPDFAsImage(filepath)
				if err != nil {
					return nil, fmt.Errorf("error processing PDF as image: %v", err)
				}
			} else {
				return nil, fmt.Errorf("error processing PDF: %v", err)
			}
		}

	case ".doc", ".docx":
		log.Printf("Starting DOCX extraction")
		text, images, err = s.fileService.ExtractTextFromDOCX(filepath)
		if err != nil {
			log.Printf("Error processing DOCX: %v", err)
			return nil, fmt.Errorf("error processing DOCX: %v", err)
		}
		log.Printf("DOCX extraction complete. Text length: %d, Images found: %d", len(text), len(images))

	case ".xlsx", ".xls", ".csv":
		log.Printf("Starting spreadsheet extraction")
		text, images, err = s.fileService.ExtractTextFromFile(filepath)
		if err != nil {
			log.Printf("Error processing spreadsheet: %v", err)
			return nil, fmt.Errorf("error processing spreadsheet: %v", err)
		}
		log.Printf("Spreadsheet extraction complete. Text length: %d", len(text))

	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		log.Printf("Starting image analysis")
		imgData, err := s.fileService.ReadImage(filepath)
		if err != nil {
			log.Printf("Error reading image: %v", err)
			return nil, fmt.Errorf("error reading image: %v", err)
		}
		text, err = s.imageService.ExtractInformationFromImage(imgData)
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
	if len(images) > 0 {
		log.Printf("Processing %d images", len(images))
		result.ImageAnalysis = make(map[string]string)
		for _, img := range images {
			imgData, err := s.fileService.ReadImage(img)
			if err != nil {
				log.Printf("Warning: Could not read image %s: %v", img, err)
				continue
			}

			analysis, err := s.imageService.ExtractInformationFromImage(imgData)
			if err != nil {
				log.Printf("Warning: Could not analyze image %s: %v", img, err)
				continue
			}

			result.ImageAnalysis[img] = analysis
		}
	}

	result.Text = text
	result.Images = images

	return result, nil
}

func (s *ProcessorService) processPDFAsImage(filepath string) (string, []string, error) {
	log.Printf("Processing PDF as image: %s", filepath)

	// Convert PDF to images and process each page
	imgData, err := s.fileService.ReadImage(filepath)
	if err != nil {
		return "", nil, fmt.Errorf("error converting PDF to images: %v", err)
	}
	text, err := s.imageService.ExtractInformationFromImage(imgData)
	if err != nil {
		return "", nil, fmt.Errorf("error analyzing image: %v", err)
	}
	return text, nil, nil
}

func (s *ProcessorService) ProcessURL(url string) (*models.ProcessingResult, error) {
	log.Printf("Processing URL: %s", url)

	// Extract text from URL
	text, headings, err := s.fileService.ExtractTextFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("error processing URL: %v", err)
	}

	// Combine text with headings for processing
	var combinedText strings.Builder
	for _, heading := range headings {
		combinedText.WriteString(heading)
		combinedText.WriteString("\n")
	}
	combinedText.WriteString(text)

	result := &models.ProcessingResult{
		Text: combinedText.String(),
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
