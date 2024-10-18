package main

import (
	"nerde_yenir/controller"
	"nerde_yenir/db"
	"nerde_yenir/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	post := db.UserData(db.Client, "post")

	postService := controller.NewPostService(post)
	postService.Start()

	router := gin.New()

	corsConfig := cors.Config{
		AllowAllOrigins: true,                                                // Tüm origin'lere izin verir, daha güvenli bir yapı için bu kısmı güncelleyin
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},            // İzin verilen HTTP yöntemleri
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"}, // İzin verilen başlıklar
	}

	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())

	routers.UserRoutes(router)
	router.POST("/addpost", controller.AddPost(postService))
	router.POST("/add-comment/:postid", controller.AddComment(postService))
	router.GET("/post/:postid", controller.GetPost(postService))
	router.GET("/sse", controller.SSEGetPost(postService))
	router.GET("/postlike", controller.PostLike(postService))
	router.POST("/add-image", controller.AddImage())

	//post service
	router.GET("/:user_name/profil", controller.GetProfilDetails(postService))
	router.GET("/:user_name/profil/:post_id", controller.GetProfilPostDetails(postService))
	router.GET("/:user_name/:post_id/del", controller.DelProfilDetail(postService))

	router.Run()
}
