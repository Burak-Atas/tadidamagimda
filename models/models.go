package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // MongoDB ObjectID
	UserId    string             `bson:"user_id"`       // Kullanıcı ID
	FirstName string             `bson:"first_name"`    // Kullanıcının adı
	LastName  string             `bson:"last_name"`     // Kullanıcının soyadı
	Email     string             `bson:"email"`         // Kullanıcının e-postası
	Password  string             `bson:"password"`      // Kullanıcının şifresi (şifrelenmiş)
	UserType  string             `bson:"user_type"`     // Kullanıcının türü (admin, user, vb.)
	Token     string             `bson:"token"`         // JWT veya diğer kimlik doğrulama tokeni
	CreatedAt time.Time          `bson:"created_at"`    // Oluşturulma zamanı
	UpdatedAt time.Time          `bson:"updated_at"`    // Güncellenme zamanı
}
type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // MongoDB ObjectID
	SenderId  string             `bson:"sender_id"`     // Göndericinin ID'si
	PostId    string             `bson:"post_id"`       // Gönderinin ID'si
	ImageUrl  string             `bson:"image_url"`     // Resim URL'si
	Text      string             `bson:"text"`          // Gönderinin içeriği
	CreatedAt time.Time          `bson:"created_at"`    // Oluşturulma zamanı
	Latitude  float32            `bson:"latitude"`      // Enlem bilgisi
	Langitude float32            `bson:"langitude"`     // Boylam bilgisi
	CountLike int                `bson:"count_like"`    // Beğeni sayısı
	Comments  []Comment          `bson:"comments"`      // Yorumlar
}

type Comment struct {
	PostId     string `bson:"post_id"`    // Gönderinin ID'si
	Comment_Id string `bson:"comment_id"` // Yorumun ID'si
	SenderId   string `bson:"sender_id"`  // Yorum gönderenin ID'si
	Text       string `bson:"text"`       // Yorumun içeriği
	Rating     int    `bson:"rating"`     // Yorumdaki puan
}

type RegionalFood struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"` // MongoDB ObjectID
	RegionalFood string             `bson:"regional_food"` // Yöresel yemek ismi
	Name         string             `bson:"name"`          // Yemeğin adı
	ImageUrl     string             `bson:"image_url"`     // Görsel URL'si
	Text         string             `bson:"text"`          // Açıklama metni
	Resturant    []Resturant        `bson:"resturant"`     // Restoranlar
}

type Resturant struct {
	ResturanId string
	Name       string
	Location   string
	//Menu []Menu
	//Rating int
}

type City struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Food      []FameousFood      `bson:"food"`
	ByAddCity string
	CityID    string
}

type FameousFood struct {
	Name        string   `bson:"name"`
	Ingredients []string `bson:"ingredients"`
	Description string   `bson:"description,omitempty"`
	ImageURL    string
}
