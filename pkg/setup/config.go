package setup

import "github.com/tahardi/bearclave/internal/setup"

type Config = setup.Config
type IPC = setup.IPC
type Server = setup.Server
type Proxy = setup.Proxy

func LoadConfig(configFile string) (*Config, error) {
	return setup.LoadConfig(configFile)
}
