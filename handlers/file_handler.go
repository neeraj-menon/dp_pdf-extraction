package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"file_processor/models"
	"file_processor/services"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	processor *services.ProcessorService
	gemini    *services.GeminiService
	rules     *services.RuleService
}

func NewFileHandler(processor *services.ProcessorService, gemini *services.GeminiService, rules *services.RuleService) *FileHandler {
	return &FileHandler{
		processor: processor,
		gemini:    gemini,
		rules:     rules,
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

	log.Printf("Received file: %s, Size: %d bytes", file.Filename, file.Size)

	// Get file extension
	ext := filepath.Ext(file.Filename)
	log.Printf("File extension: %s", ext)

	// Save file temporarily
	tempFile := filepath.Join(os.TempDir(), file.Filename)
	log.Printf("Saving file temporarily to: %s", tempFile)
	if err := c.SaveUploadedFile(file, tempFile); err != nil {
		log.Printf("Error saving file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}
	defer os.Remove(tempFile)

	// Get initial processing result
	result, err := h.processor.ProcessFile(tempFile, ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Process with Gemini
	structuredResp, err := h.gemini.ProcessDocument(c.Request.Context(), result.Text, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to process with Gemini: %v", err),
		})
		return
	}

	// Handle rules or data based on the response
	if structuredResp.IsRuleset {
		if err := h.rules.AddRule(c.Request.Context(), structuredResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to add rule: %v", err),
			})
			return
		}
	} else {
		if err := h.rules.AddData(c.Request.Context(), structuredResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to add data: %v", err),
			})
			return
		}
	}

	log.Printf("Successfully processed file. Document ID: %s", structuredResp.DocumentID)
	c.JSON(http.StatusOK, structuredResp)
}

func (h *FileHandler) HandleURLProcess(c *gin.Context) {
	var input models.URLInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL format",
		})
		return
	}

	// Get initial processing result
	result, err := h.processor.ProcessURL(input.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Process with Gemini
	structuredResp, err := h.gemini.ProcessDocument(c.Request.Context(), result.Text, input.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to process with Gemini: %v", err),
		})
		return
	}

	// Handle rules or data based on the response
	if structuredResp.IsRuleset {
		if err := h.rules.AddRule(c.Request.Context(), structuredResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to add rule: %v", err),
			})
			return
		}
	} else {
		if err := h.rules.AddData(c.Request.Context(), structuredResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to add data: %v", err),
			})
			return
		}
	}

	log.Printf("Successfully processed URL. Document ID: %s", structuredResp.DocumentID)
	c.JSON(http.StatusOK, structuredResp)
}
