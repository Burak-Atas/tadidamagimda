package controller

import (
	"context"
	"fmt"
	"nerde_yenir/db"
	"nerde_yenir/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var questionnaireCollection = db.UserData(db.Client, "QuestionnaireCollection")

func GetQuestionnaire() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		cityID := c.Query("city")
		restaurantName := c.Query("restaurantName")

		var filter bson.D

		if cityID != "" || restaurantName != "" {
			if cityID != "" {
				filter = append(filter, bson.E{Key: "city", Value: cityID})
			}
			if restaurantName != "" {
				filter = append(filter, bson.E{Key: "restaurant_name", Value: restaurantName})
			}
		}

		var results []models.RecipeFeedback
		var err error

		if len(filter) == 0 {
			cursor, err := questionnaireCollection.Find(ctx, bson.D{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questionnaires"})
				return
			}
			defer cursor.Close(ctx)

			for cursor.Next(ctx) {
				var recipeFeedback models.RecipeFeedback
				if err := cursor.Decode(&recipeFeedback); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode questionnaire"})
					return
				}
				results = append(results, recipeFeedback)
			}

		} else {
			err = questionnaireCollection.FindOne(ctx, filter).Decode(&results)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusNotFound, gin.H{"message": "No questionnaire found"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questionnaire"})
				}
				return
			}
		}

		c.JSON(http.StatusOK, results)
	}
}

func PostQuestionnaire() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(403, gin.H{
				"error": errorMessageLoggedIn,
			})
			return
		}

		var recipe models.RecipeFeedback

		/*
			image, err := c.FormFile("image")
			fmt.Print("image", image.Filename)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
				return
			}

			if image.Filename != "" {
				wg.Add(1)
				go func() {
					defer wg.Done() // Goroutine tamamlandığında WaitGroup'u güncelle

					uniqueID := uuid.New().String()
					extension := filepath.Ext(image.Filename)
					fileName := fmt.Sprintf("%s%s", uniqueID, extension)
					filePath := fmt.Sprintf("uploads/%s", fileName)

					// Resmi sunucuya kaydet
					if err := c.SaveUploadedFile(image, filePath); err != nil {
						// Hata durumunda loglama yapılabilir, ancak yanıt burada döndürülmez
						fmt.Printf("Failed to save image: %v\n", err)
						return
					}

					recipe.ImageURL = filePath
				}()
			}
		*/

		if err := c.BindJSON(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errorMessage,
			})
			return
		}

		recipe.ID = primitive.NewObjectID()
		recipe.RecipeID = recipe.ID.Hex()
		recipe.StartFormUserID = uid
		recipe.NegativeFeedback = 0
		recipe.PositiveFeedback = 0
		recipe.NegativeUserIDs = make([]string, 0)
		recipe.PositiveUserIDs = make([]string, 0)

		_, err := questionnaireCollection.InsertOne(ctx, recipe)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": errorrMessageInternalServer,
			})
			return
		}

		c.JSON(http.StatusOK, recipe)
	}
}

func GetPostQuestionnaireForUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(500, gin.H{
				"error": errorMessageLoggedIn,
			})
			return
		}

		userId := c.Query("user_id")
		if userId == "" {
			c.JSON(400, gin.H{
				"error": errorMessage,
			})
			return
		}

		var recipes []models.RecipeFeedback
		filter := bson.D{primitive.E{Key: "uid", Value: userId}}
		cursor, err := questionnaireCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(500, gin.H{
				"error": errorMessageFindDB,
			})
			return
		}

		for cursor.Next(ctx) {
			var recipeFeedback models.RecipeFeedback
			if err := cursor.Decode(&recipeFeedback); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode questionnaire"})
				return
			}
			recipes = append(recipes, recipeFeedback)
		}

		c.JSON(http.StatusOK, recipes)

	}
}

func VotePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(401, gin.H{ // 500 yerine 401: Unauthorized
				"error": errorMessageLoggedIn,
			})
			return
		}

		postID := c.Query("postID")
		if postID == "" {
			c.JSON(400, gin.H{
				"error": errorMessagePostID,
			})
			return
		}

		vt := c.Query("vote")

		if vt != "yes" && vt != "no" {
			c.JSON(400, gin.H{
				"error": "Hatalı değer, lütfen 1 veya 2 girin.",
			})
			return
		}

		filter := bson.D{primitive.E{Key: "recipeid", Value: postID}}
		fmt.Println("ffedback", filter)
		var update bson.D
		if vt == "yes" {
			update = bson.D{
				{Key: "$push", Value: bson.D{primitive.E{Key: "positiveuserids", Value: uid}}},
				{Key: "$inc", Value: bson.D{primitive.E{Key: "positivefeedback", Value: 1}}}, // "$inch" yerine "$inc"
			}
		} else {
			update = bson.D{
				{Key: "$push", Value: bson.D{primitive.E{Key: "negativeuserids", Value: uid}}},
				{Key: "$inc", Value: bson.D{primitive.E{Key: "negativefeedback", Value: 1}}}, // "$inch" yerine "$inc"
			}
		}
		fmt.Println("deneme", update)

		_, err := questionnaireCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Oylama şu anda yapılamıyor",
			})
			return
		}

		c.JSON(200, gin.H{"message": successMessage})
	}
}

func DelQuestionaire() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(401, gin.H{ // 401: Unauthorized
				"error": errorMessageLoggedIn,
			})
			return
		}

		postID := c.Query("postID")
		if postID == "" {
			c.JSON(400, gin.H{
				"error": errorMessagePostID,
			})
			return
		}

		filter := bson.D{
			{Key: "recipe_id", Value: postID},
			{Key: "start_from_user_id", Value: uid},
		}
		_, err := questionnaireCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(500, gin.H{ // Hata durumunda 500: Internal Server Error
				"error": errorMessageDelete,
			})
			return
		}

		c.JSON(200, gin.H{
			"message": successMessageDelete,
		})
	}
}
