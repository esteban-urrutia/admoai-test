package routes

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsCollector holds metrics data
type MetricsCollector struct {
	httpRequestsTotal   map[string]int
	httpRequestDuration map[string]float64
	startTime           time.Time
	mu                  sync.Mutex
}

var metricsCollector = &MetricsCollector{
	httpRequestsTotal:   make(map[string]int),
	httpRequestDuration: make(map[string]float64),
	startTime:           time.Now(),
}

// MetricsMiddleware collects HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Create metric key
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		status := fmt.Sprintf("%d", c.Writer.Status())

		key := fmt.Sprintf("%s_%s_%s", method, path, status)

		metricsCollector.mu.Lock() // Lock before accessing shared maps
		// Update metrics
		metricsCollector.httpRequestsTotal[key]++
		metricsCollector.httpRequestDuration[key] = duration
		metricsCollector.mu.Unlock() // Unlock after accessing shared maps
	}
}

// generatePrometheusMetrics generates metrics in Prometheus format
func generatePrometheusMetrics() string {
	metricsCollector.mu.Lock()         // Lock before accessing shared maps
	defer metricsCollector.mu.Unlock() // Ensure unlock even if panic occurs

	var metrics string

	// HTTP request total metrics
	metrics += "# HELP http_requests_total Total number of HTTP requests\n"
	metrics += "# TYPE http_requests_total counter\n"
	for key, count := range metricsCollector.httpRequestsTotal {
		metrics += fmt.Sprintf("http_requests_total{endpoint=\"%s\"} %d\n", key, count)
	}

	// HTTP request duration metrics
	metrics += "# HELP http_request_duration_seconds HTTP request duration in seconds\n"
	metrics += "# TYPE http_request_duration_seconds gauge\n"
	for key, duration := range metricsCollector.httpRequestDuration {
		metrics += fmt.Sprintf("http_request_duration_seconds{endpoint=\"%s\"} %.6f\n", key, duration)
	}

	// Go runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics += "# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use\n"
	metrics += "# TYPE go_memstats_alloc_bytes gauge\n"
	metrics += fmt.Sprintf("go_memstats_alloc_bytes %d\n", m.Alloc)

	metrics += "# HELP go_memstats_total_alloc_bytes Total number of bytes allocated, even if freed\n"
	metrics += "# TYPE go_memstats_total_alloc_bytes counter\n"
	metrics += fmt.Sprintf("go_memstats_total_alloc_bytes %d\n", m.TotalAlloc)

	metrics += "# HELP go_memstats_sys_bytes Number of bytes obtained from system\n"
	metrics += "# TYPE go_memstats_sys_bytes gauge\n"
	metrics += fmt.Sprintf("go_memstats_sys_bytes %d\n", m.Sys)

	metrics += "# HELP go_memstats_lookups_total Total number of pointer lookups\n"
	metrics += "# TYPE go_memstats_lookups_total counter\n"
	metrics += fmt.Sprintf("go_memstats_lookups_total %d\n", m.Lookups)

	metrics += "# HELP go_memstats_mallocs_total Total number of mallocs\n"
	metrics += "# TYPE go_memstats_mallocs_total counter\n"
	metrics += fmt.Sprintf("go_memstats_mallocs_total %d\n", m.Mallocs)

	metrics += "# HELP go_memstats_frees_total Total number of frees\n"
	metrics += "# TYPE go_memstats_frees_total counter\n"
	metrics += fmt.Sprintf("go_memstats_frees_total %d\n", m.Frees)

	metrics += "# HELP go_goroutines Number of goroutines that currently exist\n"
	metrics += "# TYPE go_goroutines gauge\n"
	metrics += fmt.Sprintf("go_goroutines %d\n", runtime.NumGoroutine())

	// Application uptime
	uptime := time.Since(metricsCollector.startTime).Seconds()
	metrics += "# HELP app_uptime_seconds Application uptime in seconds\n"
	metrics += "# TYPE app_uptime_seconds gauge\n"
	metrics += fmt.Sprintf("app_uptime_seconds %.2f\n", uptime)

	// Build info
	metrics += "# HELP go_info Information about the Go environment\n"
	metrics += "# TYPE go_info gauge\n"
	metrics += fmt.Sprintf("go_info{version=\"%s\"} 1\n", runtime.Version())

	return metrics
}

// Sets up the route for metrics endpoint
func Metrics(router *gin.Engine) {
	// Add metrics middleware to collect HTTP metrics
	router.Use(MetricsMiddleware())

	// Metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		metrics := generatePrometheusMetrics()
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(http.StatusOK, metrics)
	})
}
