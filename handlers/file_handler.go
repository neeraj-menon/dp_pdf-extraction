package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"file_processor/services"
)

type FileHandler struct {
	fileService *services.FileService
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) HandleFileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("Error getting file from request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only PDF files are supported",
		})
		return
	}

	log.Printf("Received file: %s, Size: %d bytes", file.Filename, file.Size)

	// Create uploads directory if it doesn't exist
	uploadsDir := filepath.Join("uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Error creating uploads directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Save file
	filePath := filepath.Join(uploadsDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("Error saving file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}

	// Process PDF
	text, imagePaths, err := h.fileService.ProcessPDF(filePath)
	if err != nil {
		log.Printf("Error processing PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process PDF",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File processed successfully",
		"text":    text,
		"pages":   len(imagePaths),
	})
}
