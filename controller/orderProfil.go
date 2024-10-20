package controller

import (
	"context"
	"nerde_yenir/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetOrderProfil(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user_name := c.Param("user_name")

		var user models.User
		filter := bson.D{primitive.E{Key: "user_name", Value: user_name}}
		err := UserCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			c.JSON(500, gin.H{
				"error": errorMessageFindDB,
			})
			return
		}

		postFilter := bson.D{primitive.E{Key: "sender_id", Value: user.UserId}}
		cursor, err := ps.PostCollection.Find(ctx, postFilter)
		if err != nil {
			c.JSON(500, gin.H{
				"error": errorMessageFindDB,
			})
			return
		}

		var post []models.Post

		if err := cursor.All(ctx, &post); err != nil {
			c.JSON(400, gin.H{
				"error": errorMessageFindDB,
			})
			return
		}

		c.JSON(200, gin.H{
			"post":       post,
			"user_name":  user.UserName,
			"bio":        user.Biography,
			"first_name": user.FirstName,
		})
	}
}
