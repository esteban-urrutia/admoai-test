package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Ad{})

	return db
}

func TestGetAdByID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Create test ad
	testAd := models.Ad{
		Title:                 "Test Ad",
		ImageUrl:              "https://example.com/image.jpg",
		Placement:             "homepage",
		Status:                "active",
		ExpirationTimeMinutes: 10,
	}
	db.Create(&testAd)

	// Test getting active ad
	req, _ := http.NewRequest("GET", "/ads/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Ad
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test Ad", response.Title)
}

func TestGetAdByIDNotFound(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Test getting non-existent ad
	req, _ := http.NewRequest("GET", "/ads/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateAd(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Create new ad
	newAd := models.Ad{
		Title:     "New Test Ad",
		ImageUrl:  "https://example.com/new-image.jpg",
		Placement: "sidebar",
	}

	jsonData, _ := json.Marshal(newAd)
	req, _ := http.NewRequest("POST", "/ads", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Ad
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "New Test Ad", response.Title)
	assert.Equal(t, "active", response.Status)
	assert.Equal(t, 10, response.ExpirationTimeMinutes) // Default value
}

func TestCreateAdWithCustomExpiration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Create new ad with custom expiration
	newAd := models.Ad{
		Title:                 "Custom Expiration Ad",
		ImageUrl:              "https://example.com/image.jpg",
		Placement:             "footer",
		ExpirationTimeMinutes: 30,
	}

	jsonData, _ := json.Marshal(newAd)
	req, _ := http.NewRequest("POST", "/ads", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Ad
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 30, response.ExpirationTimeMinutes)
}

func TestDeactivateAd(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Create test ad
	testAd := models.Ad{
		Title:                 "Test Ad",
		ImageUrl:              "https://example.com/image.jpg",
		Placement:             "homepage",
		Status:                "active",
		ExpirationTimeMinutes: 10,
	}
	db.Create(&testAd)

	// Deactivate ad
	req, _ := http.NewRequest("DELETE", "/ads/1/deactivate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify ad is deactivated
	var updatedAd models.Ad
	db.First(&updatedAd, 1)
	assert.Equal(t, "inactive", updatedAd.Status)
	assert.False(t, updatedAd.DeactivatedAt.IsZero())
}

func TestReactivateAd(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Create inactive test ad
	testAd := models.Ad{
		Title:                 "Test Ad",
		ImageUrl:              "https://example.com/image.jpg",
		Placement:             "homepage",
		Status:                "inactive",
		ExpirationTimeMinutes: 10,
	}
	db.Create(&testAd)

	// Reactivate ad
	req, _ := http.NewRequest("PUT", "/ads/1/reactivate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify ad is reactivated
	var updatedAd models.Ad
	db.First(&updatedAd, 1)
	assert.Equal(t, "active", updatedAd.Status)
	assert.True(t, updatedAd.DeactivatedAt.IsZero())
}

func TestGetAdsByPlacementAndStatus(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	// Create test ads
	ads := []models.Ad{
		{
			Title:                 "Homepage Ad 1",
			ImageUrl:              "https://example.com/image1.jpg",
			Placement:             "homepage",
			Status:                "active",
			ExpirationTimeMinutes: 10,
		},
		{
			Title:                 "Homepage Ad 2",
			ImageUrl:              "https://example.com/image2.jpg",
			Placement:             "homepage",
			Status:                "active",
			ExpirationTimeMinutes: 15,
		},
		{
			Title:                 "Sidebar Ad",
			ImageUrl:              "https://example.com/image3.jpg",
			Placement:             "sidebar",
			Status:                "active",
			ExpirationTimeMinutes: 20,
		},
	}

	for _, ad := range ads {
		db.Create(&ad)
	}

	// Test getting ads by placement and status
	req, _ := http.NewRequest("GET", "/ads?placement=homepage&status=active", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "homepage", response["placement"])
	assert.Equal(t, "active", response["status"])
	assert.Equal(t, float64(2), response["count"])
}

func TestGetAdsMissingParameters(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	tests := []struct {
		name string
		url  string
	}{
		{"missing placement", "/ads?status=active"},
		{"missing status", "/ads?placement=homepage"},
		{"missing both", "/ads"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGetAdsInvalidStatus(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	router := gin.New()
	Ads(router, db)

	req, _ := http.NewRequest("GET", "/ads?placement=homepage&status=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
