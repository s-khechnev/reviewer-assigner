package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"prod"`
	HttpServer HttpServer `yaml:"http_server"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost"`
	Port        int           `yaml:"port" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

const VarConfigPath = "CONFIG_PATH"

func GetConfig() *Config {
	configPath := os.Getenv(VarConfigPath)
	if configPath == "" {
		log.Fatalf("env variable %s not set", VarConfigPath)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("cannot read config file: %v", err)
	}

	return &config
}
