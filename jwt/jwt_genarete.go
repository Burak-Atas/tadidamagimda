package jwt

import (
	"context"
	"log"
	"nerde_yenir/db"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SECRET_KEY = os.Getenv("SECRET_LOVE")

type SignedDetails struct {
	UserId    string
	firstName string
	lastName  string
	userType  string
	jwt.StandardClaims
}

func NewSignedDetails(UserId, firstName, lastName, userType string, hour int) *SignedDetails {
	return &SignedDetails{
		UserId:    UserId,
		firstName: firstName,
		lastName:  lastName,
		userType:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(hour)).Unix(),
		},
	}
}

func (sd *SignedDetails) TokenGenerate() (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, sd).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err

	}
	return token, err
}

func ValidateToken(signedtoken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedtoken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "The Token is invalid"
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is expired"
		return
	}
	return claims, msg
}

var UserData *mongo.Collection = db.UserData(db.Client, "Users")

func UpdateAllTokens(signedtoken string, userid string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var updateobj primitive.D
	updateobj = append(updateobj, bson.E{Key: "token", Value: signedtoken})
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateobj = append(updateobj, bson.E{Key: "updatedat", Value: updated_at})
	upsert := true
	filter := bson.M{"user_id": userid}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := UserData.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updateobj},
	},
		&opt)
	defer cancel()
	if err != nil {
		log.Panic(err)
		return
	}

}
