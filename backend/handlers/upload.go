package handlers

import (
	"csv-processor/models"
	"csv-processor/services"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler handles CSV file uploads and processing
type UploadHandler struct {
	processor *services.CSVProcessor
	uploadDir string
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(processor *services.CSVProcessor, uploadDir string) *UploadHandler {
	return &UploadHandler{
		processor: processor,
		uploadDir: uploadDir,
	}
}

// UploadCSV handles CSV file upload and initiates processing
func (h *UploadHandler) UploadCSV(c *gin.Context) {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	// Validate file type
	if filepath.Ext(file.Filename) != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only CSV files are allowed",
		})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s", uuid.New().String(), file.Filename)
	filePath := filepath.Join(h.uploadDir, filename)

	// Save uploaded file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save uploaded file",
		})
		return
	}

	// Start background processing
	jobID := h.processor.ProcessCSVAsync(filePath)

	// Return job ID for status tracking
	response := models.UploadResponse{
		JobID:   jobID,
		Message: "File uploaded successfully. Processing started.",
	}

	c.JSON(http.StatusAccepted, response)
}

// GetJobStatus returns the current status of a processing job
func (h *UploadHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	
	result, exists := h.processor.GetJobStatus(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DownloadResult handles downloading processed CSV files
func (h *UploadHandler) DownloadResult(c *gin.Context) {
	filename := c.Param("filename")
	
	// Security: validate filename to prevent path traversal
	if filepath.Base(filename) != filename {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	filePath := h.processor.GetResultFilePath(filename)
	
	// Check if file exists
	if _, err := filepath.Abs(filePath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// Set headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")

	c.File(filePath)
}