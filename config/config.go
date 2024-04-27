package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var c *Config

type DBConfig struct {
	Driver string
	DSN    string
}

type ServerConfig struct {
	Port int
	Host string
	DB   DBConfig
}

type DownloadConfig struct {
	Path string
}

type Config struct {
	Server   ServerConfig
	Download DownloadConfig
}

func InitConfig() {
	configPath := os.Getenv("config")
	if configPath == "" {
		configPath = "config.yaml"
	}
	configByte, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(configByte, &c); err != nil {
		panic(err)
	}
	if c.Download.Path == "" {
		c.Download.Path = "./data/downloads"
		if _, err := os.Create(c.Download.Path); err != nil {
			panic(err)
		}
	}

}

func GetConfig() *Config {
	return c
}
