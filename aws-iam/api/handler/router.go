package handler

import (
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	r := gin.Default()

	r.POST("/push", PushObject)
	r.GET("/get/:key", GetObject)
	r.GET("/list", ListObjects)

	return r
}
