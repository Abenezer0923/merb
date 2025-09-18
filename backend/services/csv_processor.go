package services

import (
	"csv-processor/models"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CSVProcessor handles CSV file processing operations
type CSVProcessor struct {
	uploadDir string
	resultDir string
	jobs      map[string]*models.ProcessingResult
	jobsMutex sync.RWMutex
}

// NewCSVProcessor creates a new CSV processor instance
func NewCSVProcessor(uploadDir, resultDir string) *CSVProcessor {
	// Ensure directories exist
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(resultDir, 0755)

	return &CSVProcessor{
		uploadDir: uploadDir,
		resultDir: resultDir,
		jobs:      make(map[string]*models.ProcessingResult),
	}
}

// ProcessCSVAsync processes CSV file in background and returns job ID
func (p *CSVProcessor) ProcessCSVAsync(filePath string) string {
	jobID := uuid.New().String()
	
	result := &models.ProcessingResult{
		JobID:  jobID,
		Status: "processing",
	}
	
	p.jobsMutex.Lock()
	p.jobs[jobID] = result
	p.jobsMutex.Unlock()

	// Process in background
	go p.processCSVFile(jobID, filePath)
	
	return jobID
}

// GetJobStatus returns the current status of a processing job
func (p *CSVProcessor) GetJobStatus(jobID string) (*models.ProcessingResult, bool) {
	p.jobsMutex.RLock()
	defer p.jobsMutex.RUnlock()
	
	result, exists := p.jobs[jobID]
	return result, exists
}

// processCSVFile performs the actual CSV processing
func (p *CSVProcessor) processCSVFile(jobID, filePath string) {
	startTime := time.Now()
	
	p.jobsMutex.Lock()
	result := p.jobs[jobID]
	p.jobsMutex.Unlock()

	// Open input file
	file, err := os.Open(filePath)
	if err != nil {
		p.updateJobError(jobID, fmt.Sprintf("Failed to open file: %v", err))
		return
	}
	defer file.Close()

	// Create CSV reader with streaming configuration for large files
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3 // Expect exactly 3 fields
	reader.ReuseRecord = true  // Memory optimization: reuse record slice

	// Skip header if present
	if _, err := reader.Read(); err != nil {
		p.updateJobError(jobID, fmt.Sprintf("Failed to read header: %v", err))
		return
	}

	// Process records and aggregate with streaming (memory-efficient)
	departmentTotals := make(map[string]int)
	recordCount := 0
	batchSize := 1000 // Process in batches to manage memory

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			p.updateJobError(jobID, fmt.Sprintf("Error reading CSV: %v", err))
			return
		}

		// Parse sales number (streaming - one record at a time)
		sales, err := strconv.Atoi(record[2])
		if err != nil {
			continue // Skip invalid records
		}

		department := record[0]
		departmentTotals[department] += sales
		recordCount++

		// Memory management: periodic cleanup for very large files
		if recordCount%batchSize == 0 {
			// Update progress periodically
			p.updateJobProgress(jobID, recordCount)
		}
	}

	// Generate output file
	outputPath := filepath.Join(p.resultDir, fmt.Sprintf("result_%s.csv", jobID))
	if err := p.writeResultCSV(outputPath, departmentTotals); err != nil {
		p.updateJobError(jobID, fmt.Sprintf("Failed to write result: %v", err))
		return
	}

	// Update job status
	processingTime := time.Since(startTime)
	p.jobsMutex.Lock()
	result.Status = "completed"
	result.DownloadURL = fmt.Sprintf("/download/%s", filepath.Base(outputPath))
	result.ProcessingTime = processingTime
	result.DepartmentCount = len(departmentTotals)
	result.TotalRecords = recordCount
	p.jobsMutex.Unlock()

	// Clean up input file
	os.Remove(filePath)
}

// writeResultCSV writes the aggregated results to a CSV file
func (p *CSVProcessor) writeResultCSV(outputPath string, departmentTotals map[string]int) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Department Name", "Total Number of Sales"}); err != nil {
		return err
	}

	// Write data
	for department, total := range departmentTotals {
		record := []string{department, strconv.Itoa(total)}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// updateJobError updates job status with error
func (p *CSVProcessor) updateJobError(jobID, errorMsg string) {
	p.jobsMutex.Lock()
	defer p.jobsMutex.Unlock()
	
	if result, exists := p.jobs[jobID]; exists {
		result.Status = "error"
		result.Error = errorMsg
	}
}

// updateJobProgress updates job progress for large file processing
func (p *CSVProcessor) updateJobProgress(jobID string, recordsProcessed int) {
	p.jobsMutex.Lock()
	defer p.jobsMutex.Unlock()
	
	if result, exists := p.jobs[jobID]; exists {
		result.TotalRecords = recordsProcessed
		// Keep status as "processing" until complete
	}
}

// GetResultFilePath returns the full path to a result file
func (p *CSVProcessor) GetResultFilePath(filename string) string {
	return filepath.Join(p.resultDir, filename)
}