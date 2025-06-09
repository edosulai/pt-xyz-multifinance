package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	I18n     I18nConfig     `mapstructure:"i18n"`
}

type ServerConfig struct {
	HTTPPort int `mapstructure:"http_port"`
	GRPCPort int `mapstructure:"grpc_port"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	SecretKey  string        `mapstructure:"secret_key"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

type LoggingConfig struct {
	Level       string   `mapstructure:"level"`
	Encoding    string   `mapstructure:"encoding"`
	OutputPaths []string `mapstructure:"output_paths"`
}

type I18nConfig struct {
	DefaultLanguage    string   `mapstructure:"default_language"`
	AvailableLanguages []string `mapstructure:"available_languages"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)

	// Set replacement values from environment variables
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "root")
	viper.SetDefault("database.name", "xyz_multifinance")
	viper.SetDefault("database.ssl_mode", "disable")

	// Enable environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Map config keys to environment variables
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.username", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.ssl_mode", "DB_SSLMODE")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
