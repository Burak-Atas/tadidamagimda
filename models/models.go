package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`                // MongoDB ObjectID
	UserId         string             `bson:"user_id" json:"user_id"`       // Kullanıcı ID
	FirstName      string             `bson:"first_name" json:"first_name"` // Kullanıcının adı
	LastName       string             `bson:"last_name" json:"last_name"`   // Kullanıcının soyadı
	Email          string             `bson:"email" json:"email"`           // Kullanıcının e-postası
	UserName       string             `bson:"user_name" json:"user_name"`
	ProfilImageURL string             `json:"profil_image_url" bson:"profil_image_url"`
	Biography      string             `bson:"biography" json:"biography"`
	Password       string             `bson:"password" json:"password"` // Kullanıcının şifresi (şifrelenmiş)
	UserType       string             `bson:"user_type"`                // Kullanıcının türü (admin, user, vb.)
	Token          string             `bson:"token"`                    // JWT veya diğer kimlik doğrulama tokeni
	CreatedAt      time.Time          `bson:"created_at"`               // Oluşturulma zamanı
	UpdatedAt      time.Time          `bson:"updated_at"`               // Güncellenme zamanı
	FollowList     []string           `bosn:"follow_list" json:"follow_list"`
}

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`              // MongoDB ObjectID
	SenderId  string             `bson:"sender_id"`                  // Göndericinin ID'si
	PostID    string             `bson:"post_id" json:"post_id"`     // Gönderinin ID'si
	ImageUrl  string             `bson:"image_url" json:"image_url"` // Resim URL'si
	Text      string             `bson:"text"`                       // Gönderinin içeriği
	CreatedAt time.Time          `bson:"created_at"`                 // Oluşturulma zamanı
	Latitude  float32            `bson:"latitude"`                   // Enlem bilgisi
	Langitude float32            `bson:"langitude"`                  // Boylam bilgisi
	City      string             `json:"city"`
	Tag       []string           `json:"tags"`
	CountLike int                `bson:"count_like" json:"count_like"` // Beğeni sayısı
	Comments  []Comment          `json:"comments"`                     // Yorumlar
	Likes     []string           `json:"likes"`
}

type Comment struct {
	PostId     string `bson:"post_id" json:"post_id"`       // Gönderinin ID'si
	Comment_Id string `bson:"comment_id" json:"comment_id"` // Yorumun ID'si
	SenderId   string `bson:"sender_id" json:"sender_id"`   // Yorum gönderenin ID'si
	Text       string `bson:"text" json:"text"`             // Yorumun içeriği
	Rating     int    `bson:"rating" json:"comment_like"`   // Yorumdaki puan
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
}

type City struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	CityName  string             `bson:"city_name" json:"city_name"`
	Food      []FameousFood      `bson:"food" json:"food"`
	ByAddCity string             `json:"by_add_city"`
	CityID    string             `json:"city_id"`
}

type FameousFood struct {
	Name        string   `bson:"name" json:"name"`
	Ingredients []string `bson:"ingredients" json:"ingredients"`
	Description string   `bson:"description,omitempty" json:"description"`
	ImageURL    string   `json:"image_url"`
	ByAddFood   string   `json:"by_add_food"`
}

/*
questionnaire
*/
type RecipeFeedback struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	RecipeID         string             `json:"recipe_id"`
	RestaurantName   string             `json:"restaurant_name"`   // Restoranın adı
	City             string             `json:"city"`              // Şehir adı
	ImageURL         string             `json:"image_url"`         // Yemek görseli
	Price            string             `json:"price"`             // Fiyat
	RecipeDetails    string             `json:"recipe_details"`    // Yemek içeriği
	Text             string             `json:"text"`              // Yemek hakkında yorum
	PositiveFeedback int                `json:"positive_feedback"` // Olumlu geri bildirim sayısı
	NegativeFeedback int                `json:"negative_feedback"` // Olumsuz geri bildirim sayısı
	PositiveUserIDs  []string           `json:"positive_user_ids"` // Olumlu geri bildirimde bulunan kullanıcı ID'leri
	NegativeUserIDs  []string           `json:"negative_user_ids"` // Olumsuz geri bildirimde bulunan kullanıcı ID'leri
	StartFormUserID  string             `json:"start_from_user_id"`
}
