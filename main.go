package main

import (
	"nerde_yenir/controller"
	"nerde_yenir/db"
	"nerde_yenir/helpers"
	"nerde_yenir/middleware"
	"nerde_yenir/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	post := db.UserData(db.Client, "post")

	postService := controller.NewPostService(post)
	postService.Start()

	webConn := helpers.NewWebRTC()

	go func() {
		for {
			webConn.Connection()
		}
	}()

	router := gin.New()

	corsConfig := cors.Config{
		AllowAllOrigins: true,                                                // Tüm origin'lere izin verir, daha güvenli bir yapı için bu kısmı güncelleyin
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},            // İzin verilen HTTP yöntemleri
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"}, // İzin verilen başlıklar
	}

	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())
	router.Static("/static", "./static")

	router.GET("/webrtc", controller.Connection(webConn))

	routers.UserRoutes(router)
	router.GET("/sse", controller.SSEGetPost(postService))
	router.GET("/get_recipe_feed_back", controller.GetQuestionnaire())
	router.Use(middleware.Authentication())
	router.POST("/addpost", controller.AddPost(postService))

	router.POST("/add-comment/:postid", controller.AddComment(postService))
	router.DELETE("/delete-commet/:postid/:commentid", controller.DeleteComment(postService))

	router.GET("/postlike", controller.PostLike(postService))
	router.GET("/removepostlike", controller.RemovePostLike(postService))
	router.POST("/add-image", controller.AddImage())
	router.GET("/post/:postid", controller.GetPost(postService))

	//post service
	router.GET("/:user_name/profil", controller.GetProfilPostDetails(postService))
	router.GET("/:user_name/profil/:post_id", controller.GetProfilPostDetails(postService))
	router.GET("/:user_name/:post_id/del", controller.DelProfilDetail(postService))
	router.POST("/:user_name/profil/update", controller.UpdateProfil(postService))
	router.POST("/:user_name/profil/add-image", controller.AddProfilImage())

	//order profil  service
	router.GET("/profil/:user_name", controller.GetOrderProfil(postService))

	//anket services
	router.POST("/post_recipe_feed_back", controller.PostQuestionnaire())

	router.GET("/vote_post", controller.VotePost())

	//web rtc

	router.Run()
}
