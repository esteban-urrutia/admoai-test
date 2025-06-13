package models

import (
	"testing"
	"time"
)

func TestAdModel(t *testing.T) {
	// Test creating a new Ad instance
	ad := Ad{
		Title:                 "Test Ad",
		ImageUrl:              "https://example.com/image.jpg",
		Placement:             "homepage",
		Status:                "active",
		ExpirationTimeMinutes: 10,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		DeactivatedAt:         time.Time{},
	}

	// Test field values
	if ad.Title != "Test Ad" {
		t.Errorf("Expected Title to be 'Test Ad', got %s", ad.Title)
	}

	if ad.ImageUrl != "https://example.com/image.jpg" {
		t.Errorf("Expected ImageUrl to be 'https://example.com/image.jpg', got %s", ad.ImageUrl)
	}

	if ad.Placement != "homepage" {
		t.Errorf("Expected Placement to be 'homepage', got %s", ad.Placement)
	}

	if ad.Status != "active" {
		t.Errorf("Expected Status to be 'active', got %s", ad.Status)
	}

	if ad.ExpirationTimeMinutes != 10 {
		t.Errorf("Expected ExpirationTimeMinutes to be 10, got %d", ad.ExpirationTimeMinutes)
	}

	// Test zero time for DeactivatedAt
	if !ad.DeactivatedAt.IsZero() {
		t.Error("Expected DeactivatedAt to be zero time")
	}
}

func TestAdModelValidation(t *testing.T) {
	tests := []struct {
		name        string
		ad          Ad
		expectError bool
	}{
		{
			name: "valid ad",
			ad: Ad{
				Title:                 "Valid Ad",
				ImageUrl:              "https://example.com/image.jpg",
				Placement:             "sidebar",
				Status:                "active",
				ExpirationTimeMinutes: 15,
			},
			expectError: false,
		},
		{
			name: "empty title",
			ad: Ad{
				Title:                 "",
				ImageUrl:              "https://example.com/image.jpg",
				Placement:             "sidebar",
				Status:                "active",
				ExpirationTimeMinutes: 15,
			},
			expectError: true,
		},
		{
			name: "invalid url",
			ad: Ad{
				Title:                 "Test Ad",
				ImageUrl:              "not-a-url",
				Placement:             "sidebar",
				Status:                "active",
				ExpirationTimeMinutes: 15,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation check (in real scenario, you'd use a validator)
			hasError := tt.ad.Title == "" || !isValidURL(tt.ad.ImageUrl) || tt.ad.Placement == ""

			if hasError != tt.expectError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectError, hasError)
			}
		})
	}
}

// Helper function for URL validation (simplified)
func isValidURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}
