package main

import (
	"database/sql"
	"log"
	"models"
	"os"
	"os/signal"
	"routes"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initializeDB() (*gorm.DB, *sql.DB) {
	// define database connection string for SQLite
	dbPath := "database.db" // SQLite database file

	// get a database handler
	gormDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// get underlying *sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get database object: %v", err)
	}

	// check database connection
	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// auto-migrate models
	err = gormDB.AutoMigrate(
		&models.Ad{},
	)
	if err != nil {
		log.Fatalf("Error during auto-migration: %v", err)
	}
	log.Println("Auto-migration completed")

	return gormDB, sqlDB
}

func closeDatabaseConnection(sqlDB *sql.DB) {
	err := sqlDB.Close()
	if err != nil {
		log.Printf("--- Error closing database connection: %v", err)
		return
	}
	log.Println("---> database connection closed")
}

func gracefulShutdown(sqlDB *sql.DB) {
	log.Println("--- Initiating graceful shutdown... ---")

	closeDatabaseConnection(sqlDB)

	log.Println("--- Graceful shutdown completed")
}

func setupGracefulShutdown(sqlDB *sql.DB) {
	// Handle panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("\nPanic recovered: %v\n", r)
			gracefulShutdown(sqlDB)
			os.Exit(1)
		}
	}()

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register the channel to receive shutdown related signals
	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGQUIT, // Quit signal
		syscall.SIGHUP,  // Hangup signal
	)

	// Start a goroutine to handle signals
	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal: %v\n", sig)
		gracefulShutdown(sqlDB)
		os.Exit(0)
	}()
}

// Deactivates ads that have exceeded their default duration
func deactivateExpiredAds(gormDB *gorm.DB) {
    // Find all active ads
    var activeAds []models.Ad
    err := gormDB.Where("status = ?", "active").Find(&activeAds).Error
    if err != nil {
        log.Printf("Error fetching active ads: %v", err)
        return
    }

    var deactivatedCount int64
    for _, ad := range activeAds {
        // Calculate expiration time for this specific ad
        expirationTime := ad.CreatedAt.Add(time.Duration(ad.ExpirationTimeMinutes) * time.Minute)
        
        // Check if the ad has expired
        if time.Now().After(expirationTime) {
            // Deactivate this specific ad
            result := gormDB.Model(&ad).Updates(models.Ad{
                Status:        "inactive",
                DeactivatedAt: time.Now(),
            })
            
            if result.Error != nil {
                log.Printf("Error deactivating ad ID %d: %v", ad.ID, result.Error)
                continue
            }
            
            deactivatedCount += result.RowsAffected
        }
    }

    if deactivatedCount > 0 {
        log.Printf("Deactivated %d expired ads based on their individual expiration times", deactivatedCount)
    }
}

func main() {
	// initialize database
	gormDB, sqlDB := initializeDB()

	// Setup graceful shutdown handling
	setupGracefulShutdown(sqlDB)

	// initialize cron jobs
	c := cron.New()
	c.AddFunc("* * * * *", func() {
		deactivateExpiredAds(gormDB)
	})
	c.Start()

	// Initialize Gin router
	router := gin.Default()

	// Setup Prometheus metrics
	// This middleware will collect metrics for HTTP requests
	routes.Metrics(router)

	// Define Logs route
	// This route will handle downloading logs from the Docker container
	routes.Logs(router)

	// Define ads routes
	//
	routes.Ads(router, gormDB)

	// Start the server
	log.Println("Starting server on :8080")
	router.Run(":8080")
}
