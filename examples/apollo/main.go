package main

import (
	"fmt"

	"github.com/helin0815/gowebmonitor/pkg/apollo"
	"github.com/spf13/viper"
)

func main() {
	lsego.ConfigCenterBoot(apollo.NewApollo("servicemesh-api", "dev"))

	fmt.Println(viper.AllSettings())
	fmt.Println(viper.GetString("idaas.url"))
}
