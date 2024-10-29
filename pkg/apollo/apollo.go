package apollo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	agl "github.com/apolloconfig/agollo/v4"
	"os"

	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/helin0815/gowebmonitor/pkg/log"
	"github.com/spf13/viper"
)

var (
	defaultConfigServerURLTemplate = "http://github.%s.com"
)

type ApolloConfig struct {
	*config.AppConfig
}

// Deprecated: NewApollo has been replaced by NewApolloWithMeta.
func NewApollo(appid, env string) ConfigCenter {
	return &ApolloConfig{&config.AppConfig{
		AppID:         appid,
		Secret:        os.Getenv("APOLLO_ACCESS_KEY_SECRET"),
		IP:            fmt.Sprintf(defaultConfigServerURLTemplate, env),
		Cluster:       env,
		NamespaceName: storage.GetDefaultNamespace(),
		MustStart:     true,
	}}
}

func NewApolloWithMeta() (ConfigCenter, error) {
	appid := os.Getenv("APP_ID")
	if appid == "" {
		return nil, errors.New("appid for apollo was not found")
	}
	address := os.Getenv("APOLLO_META")
	if address == "" {
		return nil, errors.New("address for apollo was not found")
	}
	cluster := os.Getenv("IDC")
	if cluster == "" {
		return nil, errors.New("cluster for apollo was not found")
	}
	return &ApolloConfig{&config.AppConfig{
		AppID:         appid,
		Secret:        os.Getenv("APOLLO_ACCESS_KEY_SECRET"),
		IP:            address,
		Cluster:       cluster,
		NamespaceName: storage.GetDefaultNamespace(),
		MustStart:     true,
	}}, nil
}

func (c *ApolloConfig) Run() error {
	agl.SetLogger(log.SugaredLogger())
	client, err := agl.StartWithConfig(c.appConfigLoader)
	if err != nil {
		return fmt.Errorf("error starting apollo client: %v", err)
	}

	client.AddChangeListener(&configChangeListener{})
	cache := client.GetConfigCache(storage.GetDefaultNamespace())
	cfgMap := make(map[string]interface{})
	cache.Range(func(key, value interface{}) bool {
		cfgMap[key.(string)] = value
		return true
	})

	return refreshViperConfig(cfgMap)
}

func (c *ApolloConfig) appConfigLoader() (*config.AppConfig, error) {
	return c.AppConfig, nil
}

type configChangeListener struct {
}

func (c *configChangeListener) OnChange(event *storage.ChangeEvent) {
}

func (c *configChangeListener) OnNewestChange(event *storage.FullChangeEvent) {
	log.Debugf("配置发生变化：%s", event.Changes)
	if err := refreshViperConfig(event.Changes); err != nil {
		log.Errorf("Error refreshing viper configuration: %v", err)
		return
	}

	configKeys := make([]string, 0)
	for key := range event.Changes {
		configKeys = append(configKeys, key)
	}
	log.Infof("新配置已加载：%s", configKeys)
}

func refreshViperConfig(cfgMap interface{}) error {
	cfgBytes, err := json.Marshal(cfgMap)
	if err != nil {
		return fmt.Errorf("json.Marshal failed: %v", err)
	}

	viper.SetConfigType("json")
	if err := viper.ReadConfig(bytes.NewBuffer(cfgBytes)); err != nil {
		return fmt.Errorf("viper.ReadConfig Err: %v", err)
	}

	return nil
}
