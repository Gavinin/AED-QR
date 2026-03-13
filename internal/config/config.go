package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Log       LogConfig       `yaml:"log"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	Admin     AdminConfig     `yaml:"admin"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Brands    []BrandConfig   `yaml:"brands"`
}

type ServerConfig struct {
	Port   string `yaml:"port"`
	Domain string `yaml:"domain"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
}

type AdminConfig struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	OwnerPhone string `yaml:"owner_phone"`
}

type RateLimitConfig struct {
	Window        int `yaml:"window"`
	MaxRequests   int `yaml:"max_requests"`
	BlockDuration int `yaml:"block_duration"`
}

type BrandConfig struct {
	Name   string            `yaml:"name" json:"name"`
	Fields map[string]string `yaml:"fields" json:"fields"`
}

var AppConfig *Config

func LoadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	AppConfig = &Config{}
	if err := yaml.Unmarshal(data, AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
}
