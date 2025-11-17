package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Type   string `yaml:"type"`
		Sqlite struct {
			Path string `yaml:"path"`
		} `yaml:"sqlite"`
		MySQL struct {
			Host      string `yaml:"host"`
			Port      int    `yaml:"port"`
			User      string `yaml:"user"`
			Password  string `yaml:"password"`
			Database  string `yaml:"database"`
			Charset   string `yaml:"charset"`
			ParseTime bool   `yaml:"parseTime"`
			Loc       string `yaml:"loc"`
		} `yaml:"mysql"`
	} `yaml:"database"`

	Auth struct {
		JWTSecret      string `yaml:"jwt_secret"`
		AccessMinutes  int    `yaml:"access_minutes"`
		RefreshHours   int    `yaml:"refresh_hours"`
	} `yaml:"auth"`
}

func AppEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return "dev"
	}
	return strings.ToLower(env)
}

func Load(env string) (*Config, error) {
	path := fmt.Sprintf("configs/config.%s.yaml", env)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
