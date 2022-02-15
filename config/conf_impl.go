package config

import (
	"fmt"
	"os"

	"github.com/go-god/setting"
)

// ConfigOption config option
type ConfigOption struct {
	configDir  string
	configFile string
}

var (
	configDir  = "./"                      // config dir
	appEnv     = os.Getenv("app_env")      // app_env
	configFile = "app." + appEnv + ".yaml" // config file name
)

// xNew create a config interface.
func New(opts ...Option) ConfigInterface {
	c := &configImpl{}
	err := c.Load(opts...)
	if err != nil {
		panic("load config error: " + err.Error())
	}

	return c
}

type configImpl struct {
	s *setting.Setting
}

// Load load config
func (c *configImpl) Load(opts ...Option) error {
	if appEnv == "" {
		configFile = "app.yaml"
	}

	conf := &ConfigOption{
		configDir:  configDir,
		configFile: configFile,
	}

	for _, o := range opts {
		o(conf)
	}

	var err error
	c.s, err = setting.NewSetting(conf.configDir, conf.configFile)
	if err != nil {
		return fmt.Errorf("init config error: " + err.Error())
	}

	return nil
}

// IsSet is set value
func (c *configImpl) IsSet(key string) bool {
	return c.s.IsSet(key)
}

// GetValue get key to obj,obj must be a pointer
func (c *configImpl) GetValue(key string, obj interface{}) error {
	return c.s.ReadSection(key, obj)
}
