package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type FileService struct {
	tikaServerURL string
}

func NewFileService() *FileService {
	return &FileService{
		tikaServerURL: "http://localhost:9998",
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
	log.Printf("Successfully extracted %d characters of text", len(text))

	return text, []string{}, nil
}

func (s *FileService) ExtractTextFromPDF(filePath string) (string, []string, error) {
	return s.ExtractTextFromFile(filePath)
}

func (s *FileService) ExtractTextFromDOCX(filePath string) (string, []string, error) {
	return s.ExtractTextFromFile(filePath)
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

	// Extract text from various HTML elements
	c.OnHTML("p, article, section, div", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text != "" {
			log.Printf("Extracting text block: %s", truncateString(text, 100))
			textContent.WriteString(text)
			textContent.WriteString("\n\n")
		}
	})

	// Extract headings separately
	c.OnHTML("h1, h2, h3, h4, h5, h6", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text != "" {
			log.Printf("Extracting heading: %s", text)
			headings = append(headings, text)
			// Also add headings to main content
			textContent.WriteString(text)
			textContent.WriteString("\n\n")
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
		log.Printf("No text content found on webpage: %s", targetURL)
		return "", nil, fmt.Errorf("no text content found on webpage: %s", targetURL)
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

func (s *FileService) IsImageFile(filePath string) bool {
	ext := strings.ToLower(filePath)
	return strings.HasSuffix(ext, ".jpg") ||
		strings.HasSuffix(ext, ".jpeg") ||
		strings.HasSuffix(ext, ".png") ||
		strings.HasSuffix(ext, ".gif") ||
		strings.HasSuffix(ext, ".bmp")
}

func (s *FileService) ReadImage(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
