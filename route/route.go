package route

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"voice-dispatch/app/handler"
	"voice-dispatch/app/middleware"
	"voice-dispatch/config"
)

func Init() {
	route := gin.Default()
	route.Use(middleware.Cors())
	route.GET("/dispatch-init", handler.InitDispatch)
	route.GET("/get-dispatch", handler.GetDispatch)
	route.POST("/machine-Notify", handler.Notify)
	route.POST("/dispatch-edit", handler.DispatchEdit)

	err := route.Run(":" + strconv.Itoa(config.AppConfig.App.Port))
	if err != nil {
		return
	}
}
