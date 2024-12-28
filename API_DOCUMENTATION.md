# DocEx PDF API Documentation

## Overview
DocEx PDF is a document processing service that allows users to upload PDF documents and extract their contents. The service provides RESTful API endpoints for document management and processing.

## Base URL
```
http://localhost:8080
```

## API Endpoints

### Upload Document
Upload a PDF document for processing.

**Endpoint:** `/upload_document`
**Method:** `POST`
**Content-Type:** `multipart/form-data`

#### Request Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| file | File | Yes | The PDF file to be uploaded. Must be a valid PDF document. |

#### Response
```json
{
    "message": "File uploaded successfully",
    "filename": "example.pdf"
}
```

#### Error Responses
- **400 Bad Request**: When the file is missing or invalid
```json
{
    "error": "No file provided"
}
```
- **500 Internal Server Error**: When server encounters an error during processing
```json
{
    "error": "Error processing file"
}
```

## Technical Details

### File Size Limits
- Maximum file size: 8 MiB

### Supported File Types
- PDF documents (.pdf)

### Directory Structure
- `/uploads`: Directory for storing uploaded files
- `/temp`: Directory for temporary file processing

## Error Handling
The API uses standard HTTP status codes and returns JSON responses with detailed error messages when something goes wrong.

Common status codes:
- 200: Success
- 400: Bad Request
- 500: Internal Server Error

## Security Considerations
- File size is limited to prevent denial of service attacks
- File type validation is performed to ensure only PDF files are processed
- Temporary files are cleaned up after processing

## Rate Limiting
Currently, there are no rate limits implemented on the API endpoints.

## Future Enhancements
- Implementation of authentication and authorization
- Support for additional document formats
- Batch processing capabilities
- Document metadata extraction
- Search functionality for processed documents

---
Last Updated: December 28, 2024
