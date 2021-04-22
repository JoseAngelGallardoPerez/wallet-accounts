package config

import (
	"github.com/Confialink/wallet-pkg-env_config"
)

func Providers() []interface{} {
	return []interface{}{
		//*config.Config
		ReadConfig,
		//DbConfig
		func(c *Config) *env_config.Db {
			return c.Db
		},
	}
}
