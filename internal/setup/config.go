package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const DefaultConfigFile = "./configs/notee-config.yaml"

type Config struct {
	Platform Platform          `mapstructure:"platform"`
	IPCs     map[string]IPC    `mapstructure:"ipcs"`
	Servers  map[string]Server `mapstructure:"servers"`
	Proxy    Proxy             `mapstructure:"proxy"`
}

type IPC struct {
	CID  int `mapstructure:"cid"`
	Port int `mapstructure:"port"`
}

type Server struct {
	CID   int    `mapstructure:"cid"`
	Port  int    `mapstructure:"port"`
	Route string `mapstructure:"route"`
}

type Proxy struct {
	Port int `mapstructure:"port"`
}

func LoadConfig(configFile string) (*Config, error) {
	config := &Config{}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", configFile)
	}

	v := viper.New()
	v.SetConfigFile(configFile)

	ext := filepath.Ext(configFile)
	if ext != "" {
		v.SetConfigType(ext[1:]) // Remove the dot from the extension
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	return config, nil
}
