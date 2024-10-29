package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helin0815/gowebmonitor/pkg/log"
)

func main() {
	router := gin.New()
	router.Handle("GET", "/trace", func(c *gin.Context) {
		log.WithContext(c.Request.Context()).Infof("with trace")
		c.JSON(200, gin.H{
			"message": "Hello, world!",
		})

	})

	lse := lsego.Default()
	lse.RootHandle(router)
	lse.RegisterOnShutdown(func() {
		// shutdown something

		// sleep 10s to simulate some shutdown logic
		time.Sleep(time.Second * 10)
	})
	lse.HTTPServe(":9000")
}
