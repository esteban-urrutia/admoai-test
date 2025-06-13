# admoai-test

A Go-based advertisement management system with automatic expiration, metrics collection, and log management capabilities.

## Architecture Overview

This application is built using a modular architecture with the following components:

- **Web Framework**: Gin (HTTP router and middleware)
- **Database**: SQLite with GORM ORM
- **Cron Jobs**: Automated ad expiration and monitoring
- **Metrics**: Prometheus-compatible metrics endpoint
- **Logging**: Docker container log extraction and download
- **Containerization**: Docker with multi-stage builds

### Tech Stack Choices

- **Go 1.24**: Modern Go version with excellent performance and concurrency support
- **Gin**: Lightweight, fast HTTP framework with middleware support
- **GORM**: Feature-rich ORM with auto-migration capabilities
- **SQLite**: Embedded database for simplicity and portability
- **Docker**: Containerization for consistent deployment
- **Prometheus Format**: Industry-standard metrics format for monitoring

### Project Structure

```
├── main.go                 # Application entry point
├── models/                 # Data models and database schemas
│   ├── adModel.go         # Advertisement model definition
│   └── adModel_test.go    # Model tests
├── routes/                 # HTTP route handlers
│   ├── adsRoutes.go       # Advertisement CRUD operations
│   ├── metricsRoutes.go   # Prometheus metrics endpoint
│   ├── logsRoutes.go      # Log download functionality
│   └── *_test.go          # Route tests
├── data/                  # Database storage (mounted volume)
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Container orchestration
└── go.work               # Go workspace configuration
```

## Features

### Advertisement Management
- Create, retrieve, deactivate, and reactivate advertisements
- Automatic expiration based on configurable time limits
- Placement-based ad filtering (homepage, sidebar, footer)
- Status-based queries (active/inactive)

### Monitoring & Observability
- Prometheus-compatible metrics endpoint (`/metrics`)
- HTTP request metrics (count, duration)
- Go runtime metrics (memory, goroutines)
- Application uptime tracking
- Active ad quantity alerting

### Log Management
- Download Docker container logs as ZIP archives
- Date range filtering for log extraction
- Automatic filename sanitization

### Background Jobs
- Cron-based automatic ad expiration (runs every minute)
- Configurable active ad threshold monitoring

## Setup Instructions

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- Git

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd admoai-test
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Run locally**
   ```bash
   go run main.go
   ```

### Docker Setup

1. **Using Docker Compose (Recommended)**
   ```bash
   docker-compose up
   docker-compose down
   ```

2. **Manual Docker build**
   ```bash
   docker build -t admoai-test .
   docker run -p 8080:8080 -v $(pwd)/data:/sqliteData admoai-test
   ```

## API Endpoints

### Advertisement Management
- `POST /ads` - Create new advertisement
- `GET /ads/:id` - Get active advertisement by ID
- `GET /ads?placement=<placement>&status=<status>` - Filter ads by placement and status
- `DELETE /ads/:id/deactivate` - Deactivate advertisement
- `PUT /ads/:id/reactivate` - Reactivate advertisement

### Monitoring
- `GET /metrics` - Prometheus metrics endpoint

### Logs
- `GET /logs?startDate=<YYYY-MM-DD>&endDate=<YYYY-MM-DD>` - Download logs as ZIP

## Testing

### Run Unit Tests
```bash
# Test all packages
go test ./...

# Tests
go test -v ./models/...
go test -v ./routes/...
go test -v ./...
```

### API Testing Examples

```bash
# Create an ad
curl -X POST http://localhost:8080/ads \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Ad","image_url":"https://example.com/image.jpg","placement":"homepage","expiration_time":30}'

# Get active ads by placement
curl "http://localhost:8080/ads?placement=homepage&status=active"

# Get metrics
curl http://localhost:8080/metrics

# Download logs
curl "http://localhost:8080/logs?startDate=2024-01-01&endDate=2024-01-02" -o logs.zip
```

## Environment Configuration

### Database
- SQLite database stored in `./sqliteData/database.db`
- Auto-migration enabled for schema updates

### Cron Jobs
- Ad expiration check: Every minute (`* * * * *`)
- Active ad threshold: 10 ads (configurable in main.go)

### Container Settings
- Port: 8080
- Graceful shutdown with 1-minute timeout
- Log rotation: 250MB max size, 4 files retained

## Future Improvements

### Performance & Scalability
- [ ] Implement database connection pooling
- [ ] Add Redis caching for frequently accessed ads
- [ ] Implement database sharding for large-scale deployments
- [ ] Add rate limiting for API endpoints
- [ ] Implement async job processing for heavy operations

### Features
- [ ] User authentication and authorization
- [ ] Ad analytics and click tracking
- [ ] Image upload and storage integration
- [ ] Email notifications for ad expiration
- [ ] Bulk ad operations (create/update/delete multiple)
- [ ] Ad scheduling (start/end dates)
- [ ] Geographic targeting
- [ ] A/B testing capabilities

### Monitoring & Observability
- [ ] Structured logging with correlation IDs
- [ ] Distributed tracing with OpenTelemetry
- [ ] Health check endpoints
- [ ] Alerting integration (PagerDuty, Slack)
- [ ] Custom business metrics dashboards
- [ ] Log aggregation with ELK stack

### Security
- [ ] API key authentication
- [ ] Request/response validation middleware
- [ ] SQL injection prevention auditing
- [ ] TLS/HTTPS enforcement
- [ ] Rate limiting and DDoS protection

### Development & Operations
- [ ] CI/CD pipeline with automated testing
- [ ] Database migration system
- [ ] Environment-specific configuration
- [ ] Kubernetes deployment manifests
- [ ] Backup and disaster recovery procedures
- [ ] Load testing and performance benchmarks

### Code Quality
- [ ] Implement dependency injection
- [ ] Add integration tests
- [ ] API documentation with Swagger/OpenAPI
- [ ] Code linting and formatting enforcement
- [ ] Error handling standardization
- [ ] Context-based request handling
