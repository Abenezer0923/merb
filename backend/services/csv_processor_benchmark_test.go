package services

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkCSVProcessing tests performance with different file sizes
func BenchmarkCSVProcessing(b *testing.B) {
	// Test with different file sizes
	testSizes := []struct {
		name    string
		records int
	}{
		{"Small_1K", 1000},
		{"Medium_10K", 10000},
		{"Large_100K", 100000},
	}

	for _, size := range testSizes {
		b.Run(size.name, func(b *testing.B) {
			// Setup
			uploadDir := "./bench_uploads"
			resultDir := "./bench_results"
			defer os.RemoveAll(uploadDir)
			defer os.RemoveAll(resultDir)

			processor := NewCSVProcessor(uploadDir, resultDir)
			testFile := filepath.Join(uploadDir, fmt.Sprintf("bench_%d.csv", size.records))
			
			// Create benchmark CSV file
			createBenchmarkCSV(b, testFile, size.records)

			// Reset timer after setup
			b.ResetTimer()

			// Run benchmark
			for i := 0; i < b.N; i++ {
				jobID := processor.ProcessCSVAsync(testFile)
				
				// Wait for completion (simplified for benchmark)
				for {
					result, exists := processor.GetJobStatus(jobID)
					if exists && result.Status != "processing" {
						break
					}
				}
			}
		})
	}
}

// BenchmarkMemoryUsage tests memory efficiency
func BenchmarkMemoryUsage(b *testing.B) {
	uploadDir := "./bench_uploads"
	resultDir := "./bench_results"
	defer os.RemoveAll(uploadDir)
	defer os.RemoveAll(resultDir)

	processor := NewCSVProcessor(uploadDir, resultDir)
	testFile := filepath.Join(uploadDir, "memory_test.csv")
	
	// Create large file for memory testing
	createBenchmarkCSV(b, testFile, 50000)

	b.ResetTimer()
	b.ReportAllocs() // Report memory allocations

	for i := 0; i < b.N; i++ {
		jobID := processor.ProcessCSVAsync(testFile)
		
		// Wait for completion
		for {
			result, exists := processor.GetJobStatus(jobID)
			if exists && result.Status != "processing" {
				break
			}
		}
	}
}

// createBenchmarkCSV creates a CSV file with specified number of records
func createBenchmarkCSV(b *testing.B, filename string, recordCount int) {
	os.MkdirAll(filepath.Dir(filename), 0755)
	
	file, err := os.Create(filename)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Department Name", "Date", "Number of Sales"})

	// Write test data
	departments := []string{"Sales", "Marketing", "IT", "HR", "Finance"}
	for i := 0; i < recordCount; i++ {
		dept := departments[i%len(departments)]
		sales := fmt.Sprintf("%d", (i%1000)+1)
		date := "2024-01-01"
		
		writer.Write([]string{dept, date, sales})
	}
}