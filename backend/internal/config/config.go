package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Log      LogConfig      `mapstructure:"log"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Token    TokenConfig    `mapstructure:"token"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type TokenConfig struct {
	User  TokenRoleConfig `mapstructure:"user"`
	Admin TokenRoleConfig `mapstructure:"admin"`
}

type TokenRoleConfig struct {
	Expire      string `mapstructure:"expire"`
	RedisPrefix string `mapstructure:"redis_prefix"`
}

type LogConfig struct {
	Dir      string `mapstructure:"dir"`
	Level    string `mapstructure:"level"`
	MaxAge   int    `mapstructure:"max_age"`
	Compress bool   `mapstructure:"compress"`
}

type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	AllowHeaders []string `mapstructure:"allow_headers"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

var Cfg *Config

func LoadConfig(env string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "wecheckin")

	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("cors.allow_methods", []string{"get", "post", "put", "delete", "options"})
	viper.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})

	viper.SetDefault("log.dir", "./logs")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.compress", true)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults: %v", err)
	}

	if env != "" {
		viper.SetConfigName("config." + env)
		if err := viper.MergeInConfig(); err != nil {
			log.Printf("Warning: env config %s not found, skipping", "config."+env+".yaml")
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	Cfg = &cfg
	return &cfg, nil
}
