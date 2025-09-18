# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Test stage (optional)
FROM builder AS tester
RUN go test ./... -v

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o csv-processor .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/csv-processor .

# Create directories for uploads and results
RUN mkdir -p uploads results

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./csv-processor"]