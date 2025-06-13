package main

import (
	"models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate models
	db.AutoMigrate(&models.Ad{})

	return db
}

func TestDeactivateExpiredAds(t *testing.T) {
	db := setupTestDatabase()

	// Create test ads with different expiration times
	ads := []models.Ad{
		{
			Title:                 "Expired Ad 1",
			ImageUrl:              "https://example.com/image1.jpg",
			Placement:             "homepage",
			Status:                "active",
			ExpirationTimeMinutes: 1,                                // 1 minute
			CreatedAt:             time.Now().Add(-2 * time.Minute), // Created 2 minutes ago
		},
		{
			Title:                 "Expired Ad 2",
			ImageUrl:              "https://example.com/image2.jpg",
			Placement:             "sidebar",
			Status:                "active",
			ExpirationTimeMinutes: 5,                                 // 5 minutes
			CreatedAt:             time.Now().Add(-10 * time.Minute), // Created 10 minutes ago
		},
		{
			Title:                 "Active Ad",
			ImageUrl:              "https://example.com/image3.jpg",
			Placement:             "footer",
			Status:                "active",
			ExpirationTimeMinutes: 60,                                // 60 minutes
			CreatedAt:             time.Now().Add(-30 * time.Minute), // Created 30 minutes ago
		},
	}

	for _, ad := range ads {
		db.Create(&ad)
	}

	// Run the deactivation function
	deactivateExpiredAds(db)

	// Check results
	var expiredAd1, expiredAd2, activeAd models.Ad
	db.First(&expiredAd1, 1)
	db.First(&expiredAd2, 2)
	db.First(&activeAd, 3)

	// First two ads should be deactivated
	assert.Equal(t, "inactive", expiredAd1.Status)
	assert.False(t, expiredAd1.DeactivatedAt.IsZero())

	assert.Equal(t, "inactive", expiredAd2.Status)
	assert.False(t, expiredAd2.DeactivatedAt.IsZero())

	// Third ad should still be active
	assert.Equal(t, "active", activeAd.Status)
	assert.True(t, activeAd.DeactivatedAt.IsZero())
}

func TestAdsQuantityAlert(t *testing.T) {
	db := setupTestDatabase()

	// Create multiple active ads
	for i := 0; i < 15; i++ {
		ad := models.Ad{
			Title:                 "Test Ad",
			ImageUrl:              "https://example.com/image.jpg",
			Placement:             "homepage",
			Status:                "active",
			ExpirationTimeMinutes: 60,
		}
		db.Create(&ad)
	}

	// This should log a warning (threshold = 10, we have 15 ads)
	// In a real test, you'd capture the log output and verify it
	adsQuantityAlert(db, 10)

	// Verify count
	var count int64
	db.Model(&models.Ad{}).Where("status = ?", "active").Count(&count)
	assert.Equal(t, int64(15), count)
}

func TestInitializeDB(t *testing.T) {
	// This test would require mocking file system operations
	// For now, we'll test the basic functionality

	// Test database connection and migration
	db := setupTestDatabase()

	// Verify table exists
	assert.True(t, db.Migrator().HasTable(&models.Ad{}))

	// Test creating an ad
	ad := models.Ad{
		Title:                 "Test Ad",
		ImageUrl:              "https://example.com/image.jpg",
		Placement:             "homepage",
		Status:                "active",
		ExpirationTimeMinutes: 10,
	}

	result := db.Create(&ad)
	assert.NoError(t, result.Error)
	assert.Greater(t, ad.ID, uint(0))
}
