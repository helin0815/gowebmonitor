package lsego

import (
	"fmt"
	"log"

	"github.com/helin0815/gowebmonitor/pkg/apollo"
	"github.com/spf13/viper"
)

func ConfigCenterBoot(ccw apollo.ConfigCenter) {
	if err := ccw.Run(); err != nil {
		log.Fatalf("ConfigCenterBoot: %s", err)
	}

	fmt.Println("[ConfigCenter] support config keys:", viper.AllKeys())
}
