package controller

import (
	"context"
	"nerde_yenir/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FameousFoodService struct {
	foodCollection *mongo.Collection
}

func (ffs *FameousFoodService) AddCity() gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(400, gin.H{"error": "Kullanıcı kimliği bulunamadı"})
			return
		}

		var city models.City

		if err := c.BindJSON(&city); err != nil {
			c.JSON(400, gin.H{"error": "Geçersiz JSON formatı", "details": err.Error()})
			return
		}

		city.ID = primitive.NewObjectID()
		city.CityID = city.ID.Hex()
		city.ByAddCity = uid

		var ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := ffs.foodCollection.InsertOne(ctx, city)
		if err != nil {
			c.JSON(500, gin.H{"error": "Veritabanına eklenirken bir hata oluştu", "details": err.Error()})
			return
		}

		c.Header("Content-Type", "application/json")
		c.JSON(200, gin.H{"message": "Başarılı bir şekilde eklendi", "city": city})
	}
}

func (ffs *FameousFoodService) AddFameousFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		cityId := c.Query("cityId")
		if cityId == "" {
			c.JSON(400, gin.H{"error": "Lütfen şehir bilgisini seçiniz"})
			return
		}

		uid := c.GetString("uid")
		if uid != "" {
			c.JSON(400, gin.H{"error": "kullanıcığı kimliği bulunamadı"})
			return
		}

		var fameousFood models.FameousFood
		if err := c.BindJSON(&fameousFood); err != nil {
			c.JSON(400, gin.H{"error": "Geçersiz JSON formatı", "details": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: uid}}
		updateData := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "food", Value: bson.D{{Key: "$each", Value: fameousFood}}}}}}
		_, err := ffs.foodCollection.UpdateOne(ctx, filter, updateData)

		if err != nil {
			c.JSON(500, gin.H{"message": "ürün veritabanına eklenirken hata oluştu"})
			return
		}

		c.Header("Content-Type", "application/json")
		c.JSON(200, gin.H{"message": "Ünlü yemek başarıyla eklendi", "food": fameousFood})
	}
}
