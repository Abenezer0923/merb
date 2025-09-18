package main

import (
	"csv-processor/handlers"
	"csv-processor/middleware"
	"csv-processor/services"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize directories
	uploadDir := "./uploads"
	resultDir := "./results"

	// Initialize services
	processor := services.NewCSVProcessor(uploadDir, resultDir)
	uploadHandler := handlers.NewUploadHandler(processor, uploadDir)

	// Initialize Gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS())
	router.Use(middleware.SimpleAuth())

	// Set max multipart memory (for large file uploads)
	router.MaxMultipartMemory = 32 << 20 // 32 MiB

	// Routes
	api := router.Group("/api/v1")
	{
		api.POST("/upload", uploadHandler.UploadCSV)
		api.GET("/status/:jobId", uploadHandler.GetJobStatus)
		api.GET("/download/:filename", uploadHandler.DownloadResult)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "csv-processor",
		})
	})

	// Start server
	log.Println("Starting CSV Processor API on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}