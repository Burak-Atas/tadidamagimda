package controller

import (
	"context"
	"fmt"
	"nerde_yenir/models"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PostService struct {
	AddPost chan models.Post
	Comment chan models.Comment
	SSEChan chan models.Post

	PostCollection *mongo.Collection
}

func NewPostService(postCollection *mongo.Collection) *PostService {
	return &PostService{
		AddPost:        make(chan models.Post, 100),
		Comment:        make(chan models.Comment, 100),
		SSEChan:        make(chan models.Post),
		PostCollection: postCollection,
	}
}

var mu sync.Mutex

func (ps *PostService) Start() {
	go func() {
		for {
			select {
			case post := <-ps.AddPost:
				mu.Lock()
				Post = append([]models.Post{post}, Post...) // Yeni postu en başa ekle
				mu.Unlock()

				// SSE kanalına gönderileri gönder
				ps.SSEChan <- post // Yeni gönderileri tüm istemcilere ilet

			case comment := <-ps.Comment:
				mu.Lock()
				for i := range Post {
					if Post[i].PostId == comment.PostId {
						Post[i].Comments = append(Post[i].Comments, comment)
						break
					}
				}
				mu.Unlock()
			}
		}
	}()
}

/*
	func SSEGetPost(ps *PostService) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Connection", "Keep-Alive")
			c.Header("Cache-Control", "no-cache")

			// İlk başta mevcut gönderileri gönder
			c.SSEvent("message", gin.H{
				"data": "SSE message from server",
				"time": time.Now().Format(time.RFC3339),
				"post": Post,
			})
			c.Writer.Flush()

			for {
				select {
				case <-c.Writer.CloseNotify():
					fmt.Println("Client disconnected")
					return
				case newPosts := <-ps.SSEChan:
					c.SSEvent("message", gin.H{
						"data": "Yeni gönderi eklendi",
						"time": time.Now().Format(time.RFC3339),
						"post": newPosts,
					})
					c.Writer.Flush()
				}
			}
		}
	}
*/

type PostWithUsetDetails struct {
	UserName       string      `json:"user_name"`
	ImageUrlProfil string      `json:"image_url_profil"`
	Post           models.Post `json:"post"`
}

func SSEGetPost(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Connection", "Keep-Alive")
		c.Header("Cache-Control", "no-cache")
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var posts []models.Post

		cursor, err := ps.PostCollection.Find(ctx, bson.M{})
		if err != nil {
			c.SSEvent("error", gin.H{
				"data": "Veritabanı hatası",
				"time": time.Now().Format(time.RFC3339),
			})
			c.Writer.Flush()
			return
		}

		err = cursor.All(ctx, &posts)
		if err != nil {
			c.SSEvent("error", gin.H{
				"data": "Veritabanı hatası",
				"time": time.Now().Format(time.RFC3339),
			})
			c.Writer.Flush()
			return
		}

		var postWithUser []PostWithUsetDetails
		for _, p := range posts {
			user, _ := GetUserQueryUserID(p.SenderId)
			postWithUser = append(postWithUser, PostWithUsetDetails{
				UserName:       user.UserName,
				ImageUrlProfil: user.ProfilImageURL,
				Post:           p,
			})
		}
		c.SSEvent("message", gin.H{
			"data": "SSE message from server",
			"time": time.Now().Format(time.RFC3339),
			"post": postWithUser,
		})
		c.Writer.Flush()

		for {
			select {
			case <-c.Writer.CloseNotify():
				fmt.Println("Client disconnected")
				return
			case newPosts := <-ps.SSEChan:
				c.SSEvent("message", gin.H{
					"data": "Yeni gönderi eklendi",
					"time": time.Now().Format(time.RFC3339),
					"post": newPosts,
				})
				c.Writer.Flush()
			}
		}
	}
}

func GetUserQueryUserID(senderId string) (*models.User, error) {
	user := models.User{}
	err := UserCollection.FindOne(context.TODO(), bson.D{primitive.E{Key: "user_id", Value: senderId}}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

func AddPost(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(400, gin.H{"message": "kullanıcı kimliği saptanamadı"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var post models.Post
		if err := c.BindJSON(&post); err != nil {
			c.JSON(500, http.StatusInternalServerError)
			return
		}

		post.SenderId = uid
		post.Comments = make([]models.Comment, 0)
		post.PostId = primitive.NewObjectID().Hex()
		post.CountLike = 0
		post.CreatedAt = time.Now()

		// Kanal üzerinden post'u gönderiyoruz
		ps.AddPost <- post

		_, err := ps.PostCollection.InsertOne(ctx, post)
		if err != nil {
			c.JSON(500, gin.H{"message": "post eklenirken hata oluştu. Lütfen daha sonra tekrar deneyiniz"})
			return
		}

		c.JSON(200, gin.H{
			"message": "Post başarıyla eklendi!",
		})
	}
}

func AddComment(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(400, gin.H{"error": "Kullanıcı kimliği saptanamadı"})
			return
		}

		postId := c.Query("postID")
		if postId == "" {
			c.JSON(400, gin.H{"error": "hata oluştu"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		var comment models.Comment
		if err := c.BindJSON(&comment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz yorum verisi."})
			return
		}

		comment.Comment_Id = primitive.NewObjectID().Hex()
		comment.Rating = 0
		comment.SenderId = uid
		comment.PostId = postId

		filter := bson.D{primitive.E{Key: "_id", Value: postId}}
		updateData := bson.D{{Key: "$push", Value: bson.D{{Key: "comment", Value: bson.D{{Key: "$each", Value: comment}}}}}}
		ps.PostCollection.UpdateOne(ctx, filter, updateData)

		ps.Comment <- comment
		c.JSON(http.StatusOK, gin.H{
			"message": "Yorum başarıyla eklendi!",
		})
	}
}

func GetPost(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id := c.Query("postid")
		if id == "" {
			c.JSON(400, "hatalı url bilgisi")
			return
		}

		var post models.Post

		filter := bson.D{primitive.E{Key: "_id", Value: id}}
		err := ps.PostCollection.FindOne(ctx, filter).Decode(&post)

		if err != nil {
			c.JSON(400, gin.H{"error": "şu anda gönderiye erişilemiyor"})
			return
		}

		c.JSON(200, gin.H{"post": post})
	}
}

var Post []models.Post = []models.Post{
	{PostId: "1", Text: "Bu, gönderi 1 içeriğidir.", ImageUrl: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSRi2x8RiCqcGmMiQn455B9Jxup0QTcobH7bw&s", CreatedAt: time.Now(), CountLike: 4, Comments: make([]models.Comment, 0)},
	{PostId: "2", Text: "Bu, gönderi 2 içeriğidir.", ImageUrl: "https://picsum.photos/300", CreatedAt: time.Now(), CountLike: 4, Comments: make([]models.Comment, 0)},
	{PostId: "3", Text: "Bu, gönderi 3 içeriğidir.", ImageUrl: "https://picsum.photos/400", CreatedAt: time.Now(), CountLike: 4, Comments: make([]models.Comment, 0)},
	{PostId: "4", Text: "Bu, gönderi 4 içeriğidir.", ImageUrl: "https://picsum.photos/500", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
	{PostId: "5", Text: "Bu, gönderi 5 içeriğidir.", ImageUrl: "https://picsum.photos/200", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
	{PostId: "6", Text: "Bu, gönderi 6 içeriğidir.", ImageUrl: "https://example.com/image6.jpg", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
	{PostId: "7", Text: "Bu, gönderi 7 içeriğidir.", ImageUrl: "https://example.com/image7.jpg", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
	{PostId: "8", Text: "Bu, gönderi 8 içeriğidir.", ImageUrl: "https://example.com/image8.jpg", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
	{PostId: "9", Text: "Bu, gönderi 9 içeriğidir.", ImageUrl: "https://example.com/image9.jpg", CreatedAt: time.Now(), CountLike: 4, Comments: make([]models.Comment, 0)},
	{PostId: "10", Text: "Bu, gönderi 10 içeriğidir.", ImageUrl: "https://example.com/image10.jpg", Comments: make([]models.Comment, 0)},
	{PostId: "11", Text: "Gönderi 11", ImageUrl: "https://example.com/image11.jpg", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
	{PostId: "12", Text: "Gönderi 12", ImageUrl: "https://example.com/image12.jpg", CreatedAt: time.Now(), Comments: make([]models.Comment, 0)},
}

func AddImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No image uploaded",
			})
			return
		}
		filename := fmt.Sprintf("static/images/%s", file.Filename)

		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save image",
			})
			return
		}

		fileURL := fmt.Sprintf("http://localhost:8080/%s", filename)

		c.JSON(http.StatusOK, gin.H{
			"url": fileURL,
		})
	}
}

func PostLike(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		postID := c.Query("postID")
		if postID == "" {
			c.JSON(400, gin.H{"error": errorMessagePostID})
			return
		}

		filter := bson.D{primitive.E{Key: "postid", Value: postID}}
		update := bson.D{
			primitive.E{Key: "$inc", Value: bson.D{{Key: "CountLike", Value: 1}}},
		}

		_, err := ps.PostCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(400, gin.H{"error": "Unable to update CountLike"})
			return
		}

		c.JSON(200, gin.H{"message": "CountLike updated successfully"})

	}
}

// /:user_name/profil
func GetProfilDetails(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		fmt.Println("user", uid)
		if uid == "" {
			c.JSON(400, gin.H{"error": errorMessageUid})
			return
		}

		userName := c.Param("user_name")
		if userName == "" {
			c.JSON(400, gin.H{"error": errorMessageUid})
			return
		}

		filter := bson.D{primitive.E{Key: "sender_id", Value: uid}}
		cur, err := ps.PostCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(500, gin.H{
				"error": errorrMessageInternalServer,
			})
			return
		}

		var posts []models.Post

		err = cur.All(ctx, &posts)
		if err != nil {
			c.JSON(500, gin.H{
				"error": errorrMessageInternalServer,
			})
			return
		}

		c.JSON(200, gin.H{"message": successMessage, "post": posts})
	}
}

// /:user_name/profil/:post_id
func GetProfilPostDetails(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(400, gin.H{"error": errorMessageUid})
			return
		}

		userName := c.Param("user_name")
		if userName == "" {
			c.JSON(400, gin.H{"error": errorMessageUid})
			return
		}

		post_id := c.GetString("post_id")
		if uid == "" {
			c.JSON(400, gin.H{"error": errorMessagePostID})
			return
		}

		var post models.Post

		filter := bson.D{{Key: "post_id", Value: post_id}, {Key: "user_name", Value: userName}}

		err := ps.PostCollection.FindOne(ctx, filter).Decode(&post)
		if err != nil {
			c.JSON(400, gin.H{
				"error": errorMessageFindDB,
			})
			return
		}

		c.JSON(200, gin.H{"message": successMessage, "post": post})
	}
}

// /:user_name/:post_id/del
func DelProfilDetail(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(400, gin.H{"error": errorMessageUid})
			return
		}

		userName := c.Param("user_name")
		if userName == "" {
			c.JSON(400, gin.H{"error": errorMessageUid})
			return
		}

		post_id := c.GetString("post_id")
		if uid == "" {
			c.JSON(400, gin.H{"error": errorMessagePostID})
			return
		}

		filter := bson.D{{Key: "post_id", Value: post_id}, {Key: "user_name", Value: userName}}
		_, err := ps.PostCollection.DeleteOne(ctx, filter)

		if err != nil {
			c.JSON(400, gin.H{"error": errorMessageDelete})
			return
		}

		c.JSON(200, gin.H{"message": successMessageDelete})

	}
}
func UpdateProfil(ps *PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uid := c.GetString("uid")
		if uid == "" {
			c.JSON(400, gin.H{"error": "User ID is required."})
			return
		}

		userName := c.Param("user_name")
		if userName == "" {
			c.JSON(400, gin.H{"error": "Username is required."})
			return
		}

		var userDetails models.User
		if err := c.BindJSON(&userDetails); err != nil {
			c.JSON(400, gin.H{"message": "Invalid input."})
			return
		}

		filter := bson.M{"user_name": userDetails.UserName}

		var countDocuments, err = UserCollection.CountDocuments(ctx, filter)
		fmt.Println("kullanıcı adı sayısı", countDocuments)
		if err != nil {
			c.JSON(500, errorMessageAlredyUser)
			return
		}

		if countDocuments > 0 {
			c.JSON(400, gin.H{"error": errorMessageAlredyUser})
			return
		}

		filter = bson.M{"user_id": uid}

		update := bson.M{
			"$set": bson.M{
				"user_name":  userDetails.UserName,
				"first_name": userDetails.FirstName,
				"last_name":  userDetails.LastName,
				"biography":  userDetails.Biography,
			},
		}

		result, err := UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error updating user."})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(404, gin.H{"error": "User not found."})
			return
		}

		c.JSON(200, gin.H{"message": "User updated successfully."})
	}
}
