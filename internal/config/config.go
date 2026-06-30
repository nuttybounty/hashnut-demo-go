package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig  `mapstructure:"server"`
	Database DBConfig      `mapstructure:"database"`
	HashNut  HashNutConfig `mapstructure:"hashnut"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type HashNutConfig struct {
	TestMode bool   `mapstructure:"testMode"`
	BaseURL  string `mapstructure:"baseURL"`
}

func (d *DBConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + viper.GetString("database.port") +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" sslmode=" + d.SSLMode
}

func Load() (*Config, error) {
	path := os.Getenv("APP_CONF")
	if path == "" {
		path = "./etc/application.yaml"
	}
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	return &cfg, nil
}
