package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "github.com/z-y-x233/goSearch/api/v1"
)

func InitRouter(e *gin.Engine) {
	apiv1 := e.Group("api/v1")
	{
		apiv1.GET("/search/related/:query", v1.Related)
		apiv1.POST("/search", v1.Search)
		apiv1.POST("/put", v1.Put)

	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Header("Access-Control-Allow-Headers", "Content-Type,X-CSRF-Token, Authorization, Token,Access-Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS,PUT,DELETE,PATCH")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}
