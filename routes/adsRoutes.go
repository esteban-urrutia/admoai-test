package routes

import (
	"fmt"
	"models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Ads(router *gin.Engine, gormDB *gorm.DB) {

	// Get ad by id (only if active)
	router.GET("/ads/:id", func(c *gin.Context) {
		var ad models.Ad
		err := gormDB.Where("status = ?", "active").First(&ad, c.Param("id")).Error
		if err != nil {
			// If the ad is not found, return a 404 error
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found or inactive"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ad)
	})

	// Create a new ad
	router.POST("/ads", func(c *gin.Context) {
		var newAd models.Ad
		err := c.ShouldBindJSON(&newAd)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// Set status to active and DeactivatedAt to zero time (null)
		newAd.Status = "active"
		newAd.DeactivatedAt = time.Time{}

		// Set ExpirationTime to 10 if not provided
		if newAd.ExpirationTimeMinutes == 0 {
			newAd.ExpirationTimeMinutes = 10
		}

		err = gormDB.Create(&newAd).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusCreated, newAd)
	})

	// Soft delete an ad (deactivate)
	router.DELETE("/ads/:id/deactivate", func(c *gin.Context) {
		var ad models.Ad

		// Find the ad and check if it exists and is active
		err := gormDB.Where("status = ?", "active").First(&ad, c.Param("id")).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found or already inactive"})
				return
			}
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// Update status to inactive and set DeactivatedAt with current timestamp
		result := gormDB.Model(&ad).Updates(models.Ad{
			Status:        "inactive",
			DeactivatedAt: time.Now(),
		})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, result.Error.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"rows_affected": result.RowsAffected,
			"message":       fmt.Sprintf("Ad %s deactivated successfully", c.Param("id")),
		})
	})

	// reactivate an ad (soft delete)
	router.PUT("/ads/:id/reactivate", func(c *gin.Context) {
		var ad models.Ad

		// Find the ad and check if it exists and is inactive
		err := gormDB.Where("status = ?", "inactive").First(&ad, c.Param("id")).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found or already active"})
				return
			}
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		// Get ExpirationTimeMinutes
		var updateData models.Ad
		err = c.ShouldBindJSON(&updateData)
		if err == nil && updateData.ExpirationTimeMinutes > 0 {
			// Use provided ExpirationTimeMinutes
			ad.ExpirationTimeMinutes = updateData.ExpirationTimeMinutes
		} else if ad.ExpirationTimeMinutes == 0 {
			// Set ExpirationTimeMinutes to 10 if not provided
			ad.ExpirationTimeMinutes = 10
		}

		// Update status to active and reset DeactivatedAt to zero time (null)
		result := gormDB.Model(&ad).Updates(models.Ad{
			Status:                "active",
			DeactivatedAt:         time.Time{},
			ExpirationTimeMinutes: ad.ExpirationTimeMinutes,
		})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, result.Error.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"rows_affected": result.RowsAffected,
			"message":       fmt.Sprintf("Ad %s reactivated successfully", c.Param("id")),
		})
	})

	// Get ads by placement and status
	router.GET("/ads", func(c *gin.Context) {
		placement := c.Query("placement")
		status := c.Query("status")

		// Validate required parameters
		if placement == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "placement parameter is required"})
			return
		}

		if status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status parameter is required"})
			return
		}

		// Validate status value
		if status != "active" && status != "inactive" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be either 'active' or 'inactive'"})
			return
		}

		var ads []models.Ad
		err := gormDB.Where("placement = ? AND status = ?", placement, status).Find(&ads).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"placement": placement,
			"status":    status,
			"count":     len(ads),
			"ads":       ads,
		})
	})
}
