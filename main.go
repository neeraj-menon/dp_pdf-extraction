package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"file_processor/handlers"
	"file_processor/services"
)

func main() {
	// Initialize context
	ctx := context.Background()

	// Set Gemini API key directly
	geminiAPIKey := "AIzaSyDm5BZEE99IoGY0P2ZfthVbCPGgEoB44vg"

	// Initialize LLM client
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
	if err != nil {
		log.Fatalf("Error initializing Gemini client: %v", err)
	}
	defer client.Close()

	// Initialize services
	imageService := services.NewImageService(client)
	fileService := services.NewFileService(imageService)
	processorService := services.NewProcessorService(fileService, imageService)
	ruleService := services.NewRuleService()
	geminiService, err := services.NewGeminiService(geminiAPIKey, ruleService)
	if err != nil {
		log.Fatalf("Error initializing Gemini service: %v", err)
	}
	defer geminiService.Close()

	// Initialize handlers
	fileHandler := handlers.NewFileHandler(processorService, geminiService, ruleService)

	// Set up Gin router
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Configure routes
	router.POST("/upload", fileHandler.HandleFileUpload)
	router.POST("/process-url", fileHandler.HandleURLProcess)

	// Start server
	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
