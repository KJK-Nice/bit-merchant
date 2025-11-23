# Build stage
FROM golang:1.25.4-alpine AS builder

# Install git and templ
RUN apk add --no-cache git
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate templ files
RUN templ generate

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# Final stage
FROM alpine:latest

# Install CA certificates for external API calls (e.g. S3)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static assets
COPY --from=builder /app/static ./static
COPY --from=builder /app/assets ./assets

# Expose port
EXPOSE 8080

# Command to run
CMD ["./server"]

