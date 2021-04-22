package config

import (
	"os"

	"github.com/Confialink/wallet-pkg-env_config"
	"github.com/Confialink/wallet-pkg-env_mods"
	"github.com/inconshreveable/log15"
)

type Config struct {
	Env                                string
	Db                                 *env_config.Db
	Port                               string
	ProtobufPort                       string
	Cors                               *env_config.Cors
	IsEnabledScheduledTasksSimulations bool
}

// readConfig reads configs from ENV variables
func ReadConfig(logger log15.Logger) *Config {
	cfg := &Config{}
	cfg.Port = os.Getenv("VELMIE_WALLET_ACCOUNTS_PORT")
	cfg.ProtobufPort = os.Getenv("VELMIE_WALLET_ACCOUNTS_PROTOBUF_PORT")
	cfg.Env = env_config.Env("ENV", env_mods.Development)
	cfg.IsEnabledScheduledTasksSimulations = readScheduledTaskSimulation(cfg.Env)

	defaultConfigReader := env_config.NewReader("accounts")
	cfg.Cors = defaultConfigReader.ReadCorsConfig()
	cfg.Db = defaultConfigReader.ReadDbConfig()
	validateConfig(cfg, logger)

	return cfg
}

func validateConfig(cfg *Config, logger log15.Logger) {
	validator := env_config.NewValidator(logger)
	validator.ValidateCors(cfg.Cors, logger)
	validator.ValidateDb(cfg.Db, logger)
	validator.CriticalIfEmpty(cfg.Port, "VELMIE_WALLET_ACCOUNTS_PORT", logger)
	validator.CriticalIfEmpty(cfg.ProtobufPort, "VELMIE_WALLET_ACCOUNTS_PROTOBUF_PORT", logger)
}

// reads configs from ENV variable
func readScheduledTaskSimulation(env string) bool {
	isDevMode := env == "development" || env == "debug"
	if !isDevMode {
		return false
	}

	return os.Getenv("VELMIE_WALLET_ACCOUNTS_SCHEDULED_TASKS_SIMULATION_ENABLED") == "true"
}
