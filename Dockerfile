# Build the Go application
FROM golang:1.24-alpine AS builder

# Set up working directory
WORKDIR /app

# Copy all go module and workspace files
COPY go.mod go.work ./
COPY models/go.mod ./models/
COPY routes/go.mod ./routes/

# Copy the source code
COPY . .

# Build the application
RUN go build -o /admoai .

# Create a lightweight final image
FROM alpine:latest

# Install dependencies
RUN apk add --no-cache docker-cli
RUN apk add --no-cache build-base

# Copy the built executable from the builder stage
COPY --from=builder /admoai /admoai

# Expose the port the application runs on
EXPOSE 8080

# Run the application
CMD ["./admoai"]