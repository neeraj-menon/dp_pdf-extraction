package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type FileService struct {
	tikaServerURL string
	imageService  *ImageService
}

func NewFileService(imageService *ImageService) *FileService {
	return &FileService{
		tikaServerURL: "http://localhost:9998",
		imageService:  imageService,
	}
}

func (s *FileService) ExtractTextFromFile(filePath string) (string, []string, error) {
	log.Printf("Opening file: %s", filePath)

	// Check if file exists and is readable
	if _, err := os.Stat(filePath); err != nil {
		log.Printf("Error accessing file: %v", err)
		return "", nil, fmt.Errorf("error accessing file: %v", err)
	}

	// Read the file contents
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return "", nil, fmt.Errorf("error reading file: %v", err)
	}
	log.Printf("Successfully read %d bytes from file", len(fileContent))

	// Get file extension
	ext := strings.ToLower(strings.TrimPrefix(filePath[strings.LastIndex(filePath, "."):], "."))

	// Set appropriate content type based on file extension
	contentType := ""
	switch ext {
	case "pdf":
		contentType = "application/pdf"
	case "doc":
		contentType = "application/msword"
	case "docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "xls":
		contentType = "application/vnd.ms-excel"
	case "csv":
		contentType = "text/csv"
	default:
		return "", nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Create request to Tika server
	req, err := http.NewRequest("PUT", s.tikaServerURL+"/tika", bytes.NewReader(fileContent))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", contentType)

	// For Excel files, we want to preserve table structure
	if ext == "xlsx" || ext == "xls" || ext == "csv" {
		req.Header.Set("X-Tika-PDFextractInlineImages", "true")
		req.Header.Set("X-Tika-Preserve-Structure", "true")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to Tika server: %v", err)
		return "", nil, fmt.Errorf("error making request to Tika server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("tika server error (status %d): %s", resp.StatusCode, string(body))
	}

	extractedText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return "", nil, fmt.Errorf("error reading response: %v", err)
	}

	text := string(extractedText)
	if strings.Contains(text, "<?xml") && !strings.Contains(text, "<p>") {
		log.Printf("Tika returned only metadata without text content")
		return "", nil, fmt.Errorf("no text content found in file")
	}

	cleanedText := s.cleanText(text)
	if cleanedText == "" {
		log.Printf("No text content found after cleaning")
		return "", nil, fmt.Errorf("no text content found in file")
	}

	log.Printf("Successfully extracted %d characters of text", len(text))
	return cleanedText, nil, nil
}

func (s *FileService) ExtractTextFromPDF(filePath string) (string, []string, error) {
	log.Printf("Extracting text and images from PDF: %s", filePath)

	// First, get text from entire PDF using Tika
	tikaText, _, err := s.ExtractTextFromFile(filePath)
	if err != nil {
		// If Tika fails to extract text, proceed with image processing
		if strings.Contains(err.Error(), "no text content found") {
			log.Printf("No text found in PDF using Tika, proceeding with image processing")
			return s.processPDFWithImages(filePath)
		}
		return "", nil, err
	}

	// Clean the Tika text and check if it's empty or just whitespace/escape sequences
	cleanText := s.cleanText(tikaText)
	cleanText = strings.ReplaceAll(cleanText, "\n", "")
	cleanText = strings.ReplaceAll(cleanText, "\r", "")
	cleanText = strings.ReplaceAll(cleanText, "\t", "")

	if len(cleanText) > 0 {
		log.Printf("Found valid text content using Tika, skipping image processing")
		return tikaText, nil, nil
	}

	log.Printf("No valid text content found using Tika, attempting image processing")
	return s.processPDFWithImages(filePath)
}

func (s *FileService) processPDFWithImages(filePath string) (string, []string, error) {
	// Get absolute path for the input file
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		log.Printf("Error getting absolute path for %s: %v", filePath, err)
		return "", nil, fmt.Errorf("error getting absolute path: %v", err)
	}

	// Use temp directory in project root
	tempDir := filepath.Join("..", "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Error creating temp directory: %v", err)
		return "", nil, fmt.Errorf("error creating temp directory: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	pdfImageDir := filepath.Join(tempDir, fmt.Sprintf("pdf_images_%s", timestamp))
	if err := os.MkdirAll(pdfImageDir, 0755); err != nil {
		log.Printf("Error creating image directory: %v", err)
		return "", nil, fmt.Errorf("error creating image directory: %v", err)
	}

	// Get absolute path for the output directory to pass to Python script
	absPdfImageDir, err := filepath.Abs(pdfImageDir)
	if err != nil {
		log.Printf("Error getting absolute path for output directory: %v", err)
		return "", nil, fmt.Errorf("error getting absolute path for output directory: %v", err)
	}

	// Get path to Python script
	scriptPath := filepath.Join("..", "file_processor", "scripts", "pdf_to_image.py")
	absScriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		log.Printf("Error getting script path: %v", err)
		return "", nil, fmt.Errorf("error getting script path: %v", err)
	}

	log.Printf("Using script path: %s", absScriptPath)

	// Run Python script to convert PDF pages to images
	cmd := exec.Command("python", absScriptPath, absFilePath, absPdfImageDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error running PDF to image conversion: %v\nStderr: %s", err, stderr.String())
		return "", nil, fmt.Errorf("error converting PDF to images: %v", err)
	}

	// Get image paths from script output and validate each path
	var imagePaths []string
	imageMap := make(map[int]string) // Map to store page number -> image path
	maxPage := 0

	// First pass: collect and sort image paths
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if line == "" {
			continue
		}

		// Use the path as-is from Python script
		imagePath := strings.TrimSpace(line)
		
		// Extract page number from filename
		base := filepath.Base(imagePath)
		var pageNum int
		if _, err := fmt.Sscanf(base, "page_%d.jpg", &pageNum); err != nil {
			log.Printf("Warning: Could not parse page number from filename %s: %v", base, err)
			continue
		}

		// Read the image to check if it contains actual content
		imgData, err := s.ReadImage(imagePath)
		if err != nil {
			log.Printf("Warning: Could not read image %s: %v", imagePath, err)
			continue
		}

		// Check if the image has actual content by looking at file size
		if len(imgData) < 1000 { // Skip very small images
			log.Printf("Page %d appears to be empty or too small, skipping", pageNum)
			continue
		}

		log.Printf("Found page %d with content: %s", pageNum, imagePath)
		imageMap[pageNum] = imagePath
		if pageNum > maxPage {
			maxPage = pageNum
		}
		imagePaths = append(imagePaths, imagePath)
	}

	if len(imagePaths) == 0 {
		log.Printf("No valid content found in PDF images")
		return "", nil, fmt.Errorf("no valid content found in PDF")
	}

	// Process the pages in order
	var imageTexts []string

	for page := 1; page <= maxPage; page++ {
		imagePath, exists := imageMap[page]
		if !exists {
			continue
		}

		imgData, err := s.ReadImage(imagePath)
		if err != nil {
			log.Printf("Warning: Could not read image for page %d: %v", page, err)
			continue
		}

		log.Printf("Processing content on page %d", page)
		imageText, err := s.imageService.ExtractInformationFromImage(imgData)
		if err != nil {
			log.Printf("Warning: Could not analyze image for page %d: %v", page, err)
			continue
		}

		imageTexts = append(imageTexts, fmt.Sprintf("Page %d Content:\n%s", page, imageText))
		log.Printf("Successfully processed page %d", page)
	}

	// Combine all extracted text
	completeText := strings.Join(imageTexts, "\n\n")
	log.Printf("Successfully extracted text from %d pages", len(imageTexts))
	
	return completeText, imagePaths, nil
}

func (s *FileService) ExtractTextFromDOCX(filePath string) (string, []string, error) {
	return s.ExtractTextFromFile(filePath)
}

func (s *FileService) ExtractTextFromXLSX(filePath string) (string, []string, error) {
	return s.ExtractTextFromFile(filePath)
}

func (s *FileService) ReadImage(filePath string) ([]byte, error) {
	log.Printf("Reading image file: %s", filePath)
	
	// Verify the file exists and is accessible
	if _, err := os.Stat(filePath); err != nil {
		log.Printf("Error accessing image file: %v", err)
		return nil, fmt.Errorf("error accessing image file: %v", err)
	}

	// Read the image file
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading image file: %v", err)
		return nil, fmt.Errorf("error reading image file: %v", err)
	}

	log.Printf("Successfully read %d bytes from image file", len(imageData))
	return imageData, nil
}

func (s *FileService) ExtractTextFromURL(targetURL string) (string, []string, error) {
	log.Printf("Starting to scrape webpage: %s", targetURL)

	// Parse the URL to extract the domain
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Error parsing URL: %s, Error: %v", targetURL, err)
		return "", nil, fmt.Errorf("invalid URL: %v", err)
	}

	// Initialize collector with domain restriction and async mode
	c := colly.NewCollector(
		// Allow the specific domain and its subdomains
		colly.AllowedDomains(parsedURL.Host, strings.TrimPrefix(parsedURL.Host, "www.")),
		colly.Async(true),
	)

	// Rate limiting to be respectful to websites
	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       500 * time.Millisecond,
	})

	var textContent strings.Builder
	var headings []string

	// Extract text only from <p> tags
	c.OnHTML("p", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text != "" {
			// Clean the text
			text = s.cleanText(text)
			if text != "" {
				log.Printf("Extracting paragraph: %s", truncateString(text, 100))
				textContent.WriteString(text)
				textContent.WriteString("\n\n")
			}
		}
	})

	// Extract headings for context
	c.OnHTML("h1, h2, h3", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text != "" {
			// Clean the text
			text = s.cleanText(text)
			if text != "" {
				log.Printf("Extracting heading: %s", text)
				headings = append(headings, text)
			}
		}
	})

	// Error handling during scraping
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Scraping error on %s: %v", r.Request.URL, err)
	})

	// Status logging
	c.OnResponse(func(r *colly.Response) {
		log.Printf("Successfully accessed %s (Status: %d)", r.Request.URL, r.StatusCode)
	})

	// Visit the webpage
	err = c.Visit(targetURL)
	if err != nil {
		log.Printf("Error visiting URL: %s, Error: %v", targetURL, err)
		return "", nil, fmt.Errorf("error visiting URL: %v", err)
	}

	// Wait for all scraping goroutines to finish
	c.Wait()

	// Get the final content
	content := textContent.String()
	if content == "" {
		log.Printf("No paragraph content found on webpage: %s", targetURL)
		return "", nil, fmt.Errorf("no paragraph content found on webpage: %s", targetURL)
	}

	log.Printf("Successfully extracted text from webpage: %s", targetURL)
	return content, headings, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."

}

// cleanText removes invalid characters, escape sequences, and normalizes whitespace
func (s *FileService) cleanText(text string) string {
	// Remove common escape sequences
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	
	// Replace multiple newlines with a single newline
	re := regexp.MustCompile(`\n\s*\n+`)
	text = re.ReplaceAllString(text, "\n")
	
	// Remove non-printable characters except newline
	re = regexp.MustCompile(`[^\x20-\x7E\n]`)
	text = re.ReplaceAllString(text, "")
	
	// Replace multiple spaces with a single space
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)
	
	return text
}
