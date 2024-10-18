package routers

import (
	"nerde_yenir/controller"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incommingRouter *gin.Engine) {
	incommingRouter.POST("/login", controller.Login())
	incommingRouter.POST("/signup", controller.SignUp())
	incommingRouter.GET("/userdetails", controller.UserDetails())
}
