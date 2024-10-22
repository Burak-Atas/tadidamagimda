package controller

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddProfilImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userName := c.Param("user_name")
		if userName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User name is required",
			})
			return
		}

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error: User ID not found",
			})
			return
		}

		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No image uploaded",
			})
			return
		}

		filename := filepath.Join("static/images", file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save image",
			})
			return
		}

		fileURL := fmt.Sprintf("http://localhost:8080/static/images/%s", file.Filename)
		filter := bson.D{primitive.E{Key: "user_id", Value: uid}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "profil_image_url", Value: fileURL},
		}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update profile image URL in database",
			})
			return
		}

		// Başarılı yanıt
		c.JSON(http.StatusOK, gin.H{
			"url": fileURL,
		})
	}
}
