package routers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/junwei0117/logs-collector/api/controllers"
)

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}

func SetUpRouters() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	corsMiddleware := func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "User-Agent, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, PUT, GET, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}

	r.Use(corsMiddleware)

	apiRouter := r.Group("/api")

	transfersRouter := apiRouter.Group("/transfers")
	{
		transfersRouter.GET("", controllers.GetTransfers)
		transfersRouter.GET("/counters", controllers.GetTransfersCount)
	}

	addressesRouter := apiRouter.Group("/addresses")
	{
		addressesRouter.GET(":address", controllers.GetAddresses)
		addressesRouter.GET(":address/counters", controllers.GetAddressesCount)
	}

	return r
}
