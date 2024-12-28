package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"file_processor/models"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type RuleService struct {
	db *sql.DB
}

func NewRuleService() (*RuleService, error) {
	// Database connection parameters
	connStr := "host=localhost port=5433 user=admin password=admin123 dbname=data_platform sslmode=disable"

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return &RuleService{db: db}, nil
}

func (s *RuleService) Close() error {
	return s.db.Close()
}

func (s *RuleService) AddRule(ctx context.Context, response *models.StructuredResponse) error {
	// Check if this is actually a ruleset
	if !response.IsRuleset {
		return fmt.Errorf("error: file %s is not a ruleset", response.Filename)
	}

	// Log the rule addition attempt
	log.Printf("Attempting to add rule from document: %s (Type: %s)",
		response.Filename,
		response.FileType,
	)

	// Convert entire response to JSON string
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response to JSON: %v", err)
		return fmt.Errorf("error marshaling response: %v", err)
	}

	// Insert into Uploaded_Rulesets table with data from structured response
	query := `
		INSERT INTO data_platform.Uploaded_Rulesets (filename, extracted_content, filetype)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id int
	err = s.db.QueryRowContext(ctx, query, response.Filename, string(jsonData), response.FileType).Scan(&id)

	if err != nil {
		log.Printf("Error inserting rule: %v", err)
		return fmt.Errorf("error inserting rule: %v", err)
	}

	log.Printf("Rule added successfully with ID: %d", id)
	return nil
}

func (s *RuleService) AddData(ctx context.Context, response *models.StructuredResponse) error {
	// Check if this is a ruleset (we don't want to store rulesets as data)
	if response.IsRuleset {
		return fmt.Errorf("error: file %s is a ruleset, should not be stored as data", response.Filename)
	}

	// Log the data addition attempt
	log.Printf("Attempting to add data from document: %s (Type: %s)",
		response.Filename,
		response.FileType,
	)

	// For resume type documents, try to extract and log personal information
	if response.FileType == "resume" {
		if extractedContent, ok := response.Content.ExtractedContent["personal_information"].(map[string]interface{}); ok {
			if name, ok := extractedContent["name"].(string); ok {
				log.Printf("Processing resume for: %s", name)
			}
		}
	}

	// Convert entire response to JSON string
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response to JSON: %v", err)
		return fmt.Errorf("error marshaling response: %v", err)
	}

	// Log the JSON data being stored (for debugging)
	log.Printf("Storing JSON data: %s", string(jsonData))

	// Insert into Uploaded_Data table with data from structured response
	query := `
		INSERT INTO data_platform.Uploaded_Data (filename, extracted_content, filetype)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id int
	err = s.db.QueryRowContext(ctx, query, response.Filename, string(jsonData), response.FileType).Scan(&id)

	if err != nil {
		log.Printf("Error inserting data: %v", err)
		return fmt.Errorf("error inserting data: %v", err)
	}

	log.Printf("Data added successfully with ID: %d", id)
	return nil
}
