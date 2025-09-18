import React, { useState, useRef } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

function App() {
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);
  const [jobId, setJobId] = useState(null);
  const [status, setStatus] = useState(null);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);
  const [dragOver, setDragOver] = useState(false);
  const fileInputRef = useRef(null);
  const pollIntervalRef = useRef(null);

  const handleFileSelect = (selectedFile) => {
    if (selectedFile && selectedFile.type === 'text/csv') {
      setFile(selectedFile);
      setError(null);
    } else {
      setError('Please select a valid CSV file');
      setFile(null);
    }
  };

  const handleFileInputChange = (e) => {
    const selectedFile = e.target.files[0];
    handleFileSelect(selectedFile);
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    setDragOver(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setDragOver(false);
    const droppedFile = e.dataTransfer.files[0];
    handleFileSelect(droppedFile);
  };

  const pollJobStatus = (jobId) => {
    pollIntervalRef.current = setInterval(async () => {
      try {
        const response = await axios.get(`${API_BASE_URL}/status/${jobId}`);
        const jobStatus = response.data;
        setStatus(jobStatus);

        if (jobStatus.status === 'completed') {
          setResult(jobStatus);
          clearInterval(pollIntervalRef.current);
          setUploading(false);
        } else if (jobStatus.status === 'error') {
          setError(jobStatus.error || 'Processing failed');
          clearInterval(pollIntervalRef.current);
          setUploading(false);
        }
      } catch (err) {
        setError('Failed to check job status');
        clearInterval(pollIntervalRef.current);
        setUploading(false);
      }
    }, 1000);
  };

  const handleUpload = async () => {
    if (!file) {
      setError('Please select a file first');
      return;
    }

    setUploading(true);
    setError(null);
    setResult(null);
    setStatus(null);

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await axios.post(`${API_BASE_URL}/upload`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      const { job_id } = response.data;
      setJobId(job_id);
      pollJobStatus(job_id);
    } catch (err) {
      setError(err.response?.data?.error || 'Upload failed');
      setUploading(false);
    }
  };

  const handleDownload = () => {
    if (result && result.download_url) {
      const downloadUrl = `${API_BASE_URL}${result.download_url}`;
      window.open(downloadUrl, '_blank');
    }
  };

  const formatDuration = (duration) => {
    if (!duration) return 'N/A';
    const ms = parseInt(duration) / 1000000; // Convert nanoseconds to milliseconds
    return `${ms.toFixed(2)}ms`;
  };

  return (
    <div className="container">
      <h1>CSV Processor</h1>
      <p>Upload a CSV file with departmental sales data to get aggregated results.</p>

      {error && <div className="error">{error}</div>}

      <div
        className={`upload-area ${dragOver ? 'dragover' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept=".csv"
          onChange={handleFileInputChange}
          className="file-input"
        />
        
        {file ? (
          <div>
            <p>Selected file: <strong>{file.name}</strong></p>
            <p>Size: {(file.size / 1024).toFixed(2)} KB</p>
          </div>
        ) : (
          <div>
            <p>Drag and drop a CSV file here, or click to select</p>
            <p>Expected format: Department Name, Date (YYYY-MM-DD), Number of Sales</p>
          </div>
        )}
      </div>

      <button
        className="upload-button"
        onClick={handleUpload}
        disabled={!file || uploading}
      >
        {uploading ? 'Processing...' : 'Upload and Process'}
      </button>

      {uploading && status && (
        <div className="progress-container">
          <h3>Processing Status</h3>
          <p>Job ID: {jobId}</p>
          <p>Status: {status.status}</p>
          
          {status.status === 'processing' && (
            <div className="progress-bar">
              <div className="progress-fill" style={{ width: '50%' }}></div>
            </div>
          )}
        </div>
      )}

      {result && result.status === 'completed' && (
        <div className="result-container">
          <div className="success">
            Processing completed successfully!
          </div>
          
          <div className="metrics">
            <div className="metric-item">
              <div className="metric-value">{result.department_count}</div>
              <div className="metric-label">Departments</div>
            </div>
            <div className="metric-item">
              <div className="metric-value">{result.total_records}</div>
              <div className="metric-label">Records Processed</div>
            </div>
            <div className="metric-item">
              <div className="metric-value">{formatDuration(result.processing_time)}</div>
              <div className="metric-label">Processing Time</div>
            </div>
          </div>

          <button className="download-button" onClick={handleDownload}>
            Download Results
          </button>
        </div>
      )}
    </div>
  );
}

export default App;