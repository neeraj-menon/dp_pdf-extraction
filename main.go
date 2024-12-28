package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"file_processor/handlers"
	"file_processor/services"
)

func main() {
	// Set Gemini API key directly
	geminiAPIKey := "AIzaSyDm5BZEE99IoGY0P2ZfthVbCPGgEoB44vg"

	// Initialize services
	ruleService, err := services.NewRuleService()
	if err != nil {
		log.Fatalf("Error initializing rule service: %v", err)
	}
	defer ruleService.Close()

	geminiService, err := services.NewGeminiService(geminiAPIKey, ruleService)
	if err != nil {
		log.Fatalf("Error initializing Gemini service: %v", err)
	}
	defer geminiService.Close()

	imageService := services.NewImageService(geminiService.GetClient())
	fileService := services.NewFileService(imageService, geminiService)

	// Initialize handlers
	fileHandler := handlers.NewFileHandler(fileService)

	// Set up Gin router
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Configure routes
	router.POST("/upload_document", fileHandler.HandleFileUpload)

	// Create required directories
	requiredDirs := []string{"uploads", "temp"}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
