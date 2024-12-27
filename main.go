package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"file_processor/handlers"
	"file_processor/services"
)

func main() {
	// Initialize context
	ctx := context.Background()

	// Initialize LLM client
	llmClient, err := initializeLLMClient(ctx)
	if err != nil {
		log.Fatalf("Error initializing LLM client: %v", err)
	}
	defer llmClient.Close()

	// Set Gemini API key directly for testing
	geminiAPIKey := "AIzaSyDm5BZEE99IoGY0P2ZfthVbCPGgEoB44vg"
	if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
		geminiAPIKey = envKey
	}

	// Initialize services
	fileService := services.NewFileService()
	imageService := services.NewImageService(llmClient)
	processorService := services.NewProcessorService(fileService, imageService)

	ruleService := services.NewRuleService()

	geminiService, err := services.NewGeminiService(geminiAPIKey, ruleService)
	if err != nil {
		log.Fatalf("Error initializing Gemini service: %v", err)
	}
	defer geminiService.Close()

	// Initialize handler
	handler := handlers.NewFileHandler(processorService, geminiService, ruleService)

	// Setup router
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Define routes
	router.POST("/upload", handler.HandleFileUpload)
	router.POST("/process-url", handler.HandleURLProcess)

	// Start server
	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func initializeLLMClient(ctx context.Context) (*genai.Client, error) {
	apiKey := "AIzaSyDm5BZEE99IoGY0P2ZfthVbCPGgEoB44vg"
	return genai.NewClient(ctx, option.WithAPIKey(apiKey))
}
