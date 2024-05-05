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

func Init() {
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
		panic("download path not set")
	}

}

func GetConfig() *Config {
	return c
}

func SetConfig(conf *Config) {
	c = conf
}
