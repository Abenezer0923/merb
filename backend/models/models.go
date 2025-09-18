package models

import "time"

// SalesRecord represents a single row in the input CSV
type SalesRecord struct {
	Department string    `json:"department"`
	Date       time.Time `json:"date"`
	Sales      int       `json:"sales"`
}

// DepartmentTotal represents aggregated sales per department
type DepartmentTotal struct {
	Department string `json:"department"`
	TotalSales int    `json:"total_sales"`
}

// ProcessingResult contains the result of CSV processing
type ProcessingResult struct {
	JobID           string            `json:"job_id"`
	Status          string            `json:"status"`
	DownloadURL     string            `json:"download_url,omitempty"`
	ProcessingTime  time.Duration     `json:"processing_time,omitempty"`
	DepartmentCount int               `json:"department_count,omitempty"`
	TotalRecords    int               `json:"total_records,omitempty"`
	Error           string            `json:"error,omitempty"`
}

// UploadResponse represents the immediate response after file upload
type UploadResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}