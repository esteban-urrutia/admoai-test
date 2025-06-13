package routes

import (
	"models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Ads(router *gin.Engine, gormDB *gorm.DB) {

	// get all ads
	router.GET("/ads/all", func(c *gin.Context) {
		var ads []models.Ad
		err := gormDB.Find(&ads).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, ads)
	})

	// get a specific ad by id
	router.GET("/ads/:id", func(c *gin.Context) {
		var ad models.Ad
		err := gormDB.First(&ad, c.Param("id")).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, ad)
	})

	// create a new ad
	router.POST("/ads/add", func(c *gin.Context) {
		var newAd models.Ad
		err := c.ShouldBindJSON(&newAd)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		err = gormDB.Create(&newAd).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusCreated, newAd)
	})

}
