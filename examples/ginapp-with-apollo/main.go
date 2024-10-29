package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/helin0815/gowebmonitor/pkg/apollo"
	"github.com/spf13/viper"
)

func main() {
	lsego.ConfigCenterBoot(apollo.NewApollo("servicemesh-api", "dev"))
	router := gin.Default()
	router.Handle("GET", "/abc", func(c *gin.Context) {
		fmt.Println(viper.Get("idaas.url"))
		c.JSON(200, gin.H{
			"message": "Hello, world!",
		})
	})

	lse := lsego.New()
	lse.RootHandle(router)
	lse.RegisterOnShutdown(func() {
		// shutdown something

		// sleep 10s to simulate some shutdown logic
		time.Sleep(time.Second * 10)
	})
	lse.HTTPServe(":9000")
}
