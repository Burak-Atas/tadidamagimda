package controller

import (
	"context"
	"fmt"
	"log"
	"nerde_yenir/db"
	"nerde_yenir/helpers"
	"nerde_yenir/jwt"
	"nerde_yenir/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection *mongo.Collection = db.UserData(db.Client, "Users")

var Validate = validator.New()

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		err := UserCollection.FindOne(ctx, bson.M{"user_name": user.Email}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}
		PasswordIsValid, msg := helpers.VerifyPassword(user.Password, founduser.Password)
		defer cancel()
		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		sd := jwt.NewSignedDetails(founduser.UserId, founduser.FirstName, founduser.LastName, founduser.UserType, 23)
		token, err := sd.TokenGenerate()

		if err != nil {
			return
		}
		defer cancel()
		jwt.UpdateAllTokens(token, founduser.UserId)
		c.JSON(http.StatusOK, gin.H{"user": founduser, token: token})
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		}

		password := helpers.HashPassword(user.Password)
		user.Password = password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserId = user.ID.Hex()

		sd := jwt.NewSignedDetails(user.UserId, user.FirstName, user.LastName, user.UserType, 23)
		token, err := sd.TokenGenerate()
		if err != nil {
			c.JSON(400, gin.H{"message": "Kullanıcı oluşturulurken bir hata oluştu. Lütfen daha sonra tekrar deneyiniz"})
			return
		}

		user.Token = token
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusCreated, user)
	}
}

func UserDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		token := c.Query("token")
		if token == "" {
			c.JSON(400, gin.H{
				"error": errorMessageTokenNotFound,
			})
			return
		}

		claims, tokenerr := jwt.ValidateToken(token)
		fmt.Println(tokenerr)
		if tokenerr != "" {
			c.JSON(400, gin.H{
				"error": "lütfen tekrar giriş yapın",
			})
			return
		}

		var foundUser models.User
		filter := bson.D{primitive.E{Key: "user_id", Value: claims.UserId}}
		err := UserCollection.FindOne(ctx, filter).Decode(&foundUser)
		if err != nil {
			c.JSON(400, gin.H{
				"error": errorMessageTokenNotFound,
			})
			return
		}

		c.JSON(200, gin.H{
			"user": foundUser,
		})
	}
}
