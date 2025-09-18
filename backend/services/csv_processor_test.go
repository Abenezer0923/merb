package services

import (
	"csv-processor/models"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVProcessor_ProcessCSVAsync(t *testing.T) {
	// Setup test directories
	uploadDir := "./test_uploads"
	resultDir := "./test_results"
	defer os.RemoveAll(uploadDir)
	defer os.RemoveAll(resultDir)

	processor := NewCSVProcessor(uploadDir, resultDir)

	// Create test CSV file
	testFile := filepath.Join(uploadDir, "test.csv")
	createTestCSV(t, testFile)

	// Process CSV
	jobID := processor.ProcessCSVAsync(testFile)
	assert.NotEmpty(t, jobID)

	// Wait for processing to complete
	var result *models.ProcessingResult
	var exists bool
	for i := 0; i < 10; i++ {
		result, exists = processor.GetJobStatus(jobID)
		if exists && result.Status != "processing" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	require.True(t, exists)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, 2, result.DepartmentCount) // Sales and Marketing
	assert.Equal(t, 4, result.TotalRecords)
	assert.NotEmpty(t, result.DownloadURL)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
}

func TestCSVProcessor_GetJobStatus(t *testing.T) {
	processor := NewCSVProcessor("./test_uploads", "./test_results")

	// Test non-existent job
	result, exists := processor.GetJobStatus("non-existent")
	assert.False(t, exists)
	assert.Nil(t, result)
}

func createTestCSV(t *testing.T, filename string) {
	file, err := os.Create(filename)
	require.NoError(t, err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"Department Name", "Date", "Number of Sales"})
	require.NoError(t, err)

	// Write test data
	testData := [][]string{
		{"Sales", "2024-01-01", "100"},
		{"Marketing", "2024-01-01", "50"},
		{"Sales", "2024-01-02", "150"},
		{"Marketing", "2024-01-02", "75"},
	}

	for _, record := range testData {
		err = writer.Write(record)
		require.NoError(t, err)
	}
}