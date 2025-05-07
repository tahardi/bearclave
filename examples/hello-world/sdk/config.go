package sdk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/tahardi/bearclave"
)

const DefaultConfigFile = "Hardcoded Defaults"

type Config struct {
	Platform         Platform `mapstructure:"platform"`
	EnclaveCID       int      `mapstructure:"enclave_cid"`
	EnclavePort      int      `mapstructure:"enclave_port"`
	EnclaveProxyCID  int      `mapstructure:"enclave_proxy_cid"`
	EnclaveProxyPort int      `mapstructure:"enclave_proxy_port"`
}

func LoadConfig(configFile string) (*Config, error) {
	config := &Config{
		Platform:         Unsafe,
		EnclaveCID:       bearclave.NitroEnclaveCID,
		EnclavePort:      8082,
		EnclaveProxyCID:  bearclave.NitroEnclaveProxyCID,
		EnclaveProxyPort: 8081,
	}
	if configFile == DefaultConfigFile {
		return config, nil
	}

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
