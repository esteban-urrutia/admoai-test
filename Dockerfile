# Build the Go application
FROM golang:1.24-alpine AS builder

# Set up working directory
WORKDIR /app

# Install build dependencies for running SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Enable CGO for sqlite3
ENV CGO_ENABLED=1

# Copy all go module and workspace files
COPY go.mod go.work ./
COPY models/go.mod ./models/
COPY routes/go.mod ./routes/

# Copy the source code
COPY . .

# Build the application with CGO enabled
RUN go build -o /admoai .

# Create a lightweight final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache docker-cli sqlite

# Create the sqliteData directory for database persistence
RUN mkdir -p /sqliteData

# Copy the built executable from the builder stage
COPY --from=builder /admoai /admoai

# Expose the port the application runs on
EXPOSE 8080

# Run the application
CMD ["./admoai"]