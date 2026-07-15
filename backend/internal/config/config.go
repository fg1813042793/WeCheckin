package config

import (
	"log"
	"os"
	"strings"

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
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults: %v", err)
	}

	if env != "" {
		v.SetConfigName("config." + env)
		if err := v.MergeInConfig(); err != nil {
			log.Printf("Warning: env config %s not found, skipping", "config."+env+".yaml")
		}
	}

	bindEnv(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	Cfg = &cfg
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "8083")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "wecheckin")

	v.SetDefault("cors.allow_origins", []string{"*"})
	v.SetDefault("cors.allow_methods", []string{"get", "post", "put", "delete", "options"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})

	v.SetDefault("log.dir", "./logs")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)
}

func bindEnv(v *viper.Viper) {
	bindings := map[string]string{
		"server.port":              "WECHECKIN_SERVER_PORT",
		"server.host":              "WECHECKIN_SERVER_HOST",
		"server.mode":              "WECHECKIN_SERVER_MODE",
		"database.host":            "WECHECKIN_DATABASE_HOST",
		"database.port":            "WECHECKIN_DATABASE_PORT",
		"database.user":            "WECHECKIN_DATABASE_USER",
		"database.password":        "WECHECKIN_DATABASE_PASSWORD",
		"database.dbname":          "WECHECKIN_DATABASE_DBNAME",
		"redis.host":               "WECHECKIN_REDIS_HOST",
		"redis.port":               "WECHECKIN_REDIS_PORT",
		"redis.password":           "WECHECKIN_REDIS_PASSWORD",
		"redis.db":                 "WECHECKIN_REDIS_DB",
		"log.dir":                  "WECHECKIN_LOG_DIR",
		"log.level":                "WECHECKIN_LOG_LEVEL",
		"log.max_age":              "WECHECKIN_LOG_MAX_AGE",
		"log.compress":             "WECHECKIN_LOG_COMPRESS",
		"token.user.expire":        "WECHECKIN_TOKEN_USER_EXPIRE",
		"token.user.redis_prefix":  "WECHECKIN_TOKEN_USER_REDIS_PREFIX",
		"token.admin.expire":       "WECHECKIN_TOKEN_ADMIN_EXPIRE",
		"token.admin.redis_prefix": "WECHECKIN_TOKEN_ADMIN_REDIS_PREFIX",
		"cors.allow_origins":       "WECHECKIN_CORS_ALLOW_ORIGINS",
		"cors.allow_methods":       "WECHECKIN_CORS_ALLOW_METHODS",
		"cors.allow_headers":       "WECHECKIN_CORS_ALLOW_HEADERS",
	}
	for key, env := range bindings {
		if err := v.BindEnv(key, env); err != nil {
			log.Printf("Warning: bind env %s: %v", env, err)
		}
	}

	for key, env := range map[string]string{
		"cors.allow_origins": "WECHECKIN_CORS_ALLOW_ORIGINS",
		"cors.allow_methods": "WECHECKIN_CORS_ALLOW_METHODS",
		"cors.allow_headers": "WECHECKIN_CORS_ALLOW_HEADERS",
	} {
		if raw, ok := os.LookupEnv(env); ok {
			v.Set(key, splitCSV(raw))
		}
	}
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}
