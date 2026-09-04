package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	CORS          CORSConfig          `mapstructure:"cors"`
	Log           LogConfig           `mapstructure:"log"`
	OSS           OSSConfig           `mapstructure:"oss"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Token         TokenConfig         `mapstructure:"token"`
	ScheduledTask ScheduledTaskConfig `mapstructure:"scheduled_task"`
}

type ScheduledTaskConfig struct {
	RedisKeyPrefix           string                         `mapstructure:"redis_key_prefix"`
	WorkerCount              int                            `mapstructure:"worker_count"`
	SchedulerPollSeconds     int                            `mapstructure:"scheduler_poll_seconds"`
	SchedulerRecoverySeconds int                            `mapstructure:"scheduler_recovery_seconds"`
	WorkerPollBlockSeconds   int                            `mapstructure:"worker_poll_block_seconds"`
	WorkerHeartbeatSeconds   int                            `mapstructure:"worker_heartbeat_seconds"`
	WorkerTTLSeconds         int                            `mapstructure:"worker_ttl_seconds"`
	RecoveryTimeoutSeconds   int                            `mapstructure:"recovery_timeout_seconds"`
	MinimumSecondInterval    int                            `mapstructure:"minimum_second_interval"`
	RunRetentionDays         int                            `mapstructure:"run_retention_days"`
	LogRetentionDays         int                            `mapstructure:"log_retention_days"`
	MaxLogSegmentBytes       int                            `mapstructure:"max_log_segment_bytes"`
	MaxLogRunBytes           int                            `mapstructure:"max_log_run_bytes"`
	EnableShell              bool                           `mapstructure:"enable_shell"`
	EnableSQL                bool                           `mapstructure:"enable_sql"`
	HTTP                     HTTPPolicyConfig               `mapstructure:"http"`
	ShellCommands            map[string]ShellCommandConfig  `mapstructure:"shell_commands"`
	SQLDataSources           map[string]SQLDataSourceConfig `mapstructure:"sql_data_sources"`
}

type HTTPPolicyConfig struct {
	AllowedHosts         []string                     `mapstructure:"allowed_hosts"`
	AllowedCIDRs         []string                     `mapstructure:"allowed_cidrs"`
	Credentials          map[string]map[string]string `mapstructure:"credentials"`
	AllowPrivateNetworks bool                         `mapstructure:"allow_private_networks"`
	MaxRedirects         int                          `mapstructure:"max_redirects"`
	MaxRequestBytes      int64                        `mapstructure:"max_request_bytes"`
	MaxResponseBytes     int64                        `mapstructure:"max_response_bytes"`
}

type ShellCommandConfig struct {
	ExecutablePath  string   `mapstructure:"executable_path"`
	WorkingDir      string   `mapstructure:"working_dir"`
	AllowedEnv      []string `mapstructure:"allowed_env"`
	ArgumentPattern string   `mapstructure:"argument_pattern"`
	MaxArgs         int      `mapstructure:"max_args"`
}

type SQLDataSourceConfig struct {
	Driver             string            `mapstructure:"driver"`
	DSN                string            `mapstructure:"dsn"`
	ReadOnly           bool              `mapstructure:"read_only"`
	MaxOpenConnections int               `mapstructure:"max_open_connections"`
	Procedures         map[string]string `mapstructure:"procedures"`
}

type OSSConfig struct {
	Type    string           `mapstructure:"type"`
	Aliyun  AliyunOSSConfig  `mapstructure:"aliyun"`
	Tencent TencentOSSConfig `mapstructure:"tencent"`
	Local   LocalOSSConfig   `mapstructure:"local"`
}

type AliyunOSSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
}

type TencentOSSConfig struct {
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
}

type LocalOSSConfig struct {
	Path string `mapstructure:"path"`
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
	v.SetDefault("cors.allow_methods", []string{"get", "post", "put", "patch", "delete", "options"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})

	v.SetDefault("log.dir", "./logs")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)
	v.SetDefault("oss.type", "local")
	v.SetDefault("oss.local.path", "./uploads")
	v.SetDefault("scheduled_task.redis_key_prefix", "wecheckin")
	v.SetDefault("scheduled_task.worker_count", 4)
	v.SetDefault("scheduled_task.scheduler_poll_seconds", 2)
	v.SetDefault("scheduled_task.scheduler_recovery_seconds", 30)
	v.SetDefault("scheduled_task.worker_poll_block_seconds", 5)
	v.SetDefault("scheduled_task.worker_heartbeat_seconds", 10)
	v.SetDefault("scheduled_task.worker_ttl_seconds", 30)
	v.SetDefault("scheduled_task.recovery_timeout_seconds", 90)
	v.SetDefault("scheduled_task.minimum_second_interval", 5)
	v.SetDefault("scheduled_task.run_retention_days", 90)
	v.SetDefault("scheduled_task.log_retention_days", 30)
	v.SetDefault("scheduled_task.max_log_segment_bytes", 16384)
	v.SetDefault("scheduled_task.max_log_run_bytes", 1048576)
	v.SetDefault("scheduled_task.enable_shell", false)
	v.SetDefault("scheduled_task.enable_sql", false)
	v.SetDefault("scheduled_task.http.allow_private_networks", false)
	v.SetDefault("scheduled_task.http.max_redirects", 0)
	v.SetDefault("scheduled_task.http.max_request_bytes", 1048576)
	v.SetDefault("scheduled_task.http.max_response_bytes", 1048576)
}

func bindEnv(v *viper.Viper) {
	bindings := map[string]string{
		"server.port":                                "WECHECKIN_SERVER_PORT",
		"server.host":                                "WECHECKIN_SERVER_HOST",
		"server.mode":                                "WECHECKIN_SERVER_MODE",
		"database.host":                              "WECHECKIN_DATABASE_HOST",
		"database.port":                              "WECHECKIN_DATABASE_PORT",
		"database.user":                              "WECHECKIN_DATABASE_USER",
		"database.password":                          "WECHECKIN_DATABASE_PASSWORD",
		"database.dbname":                            "WECHECKIN_DATABASE_DBNAME",
		"redis.host":                                 "WECHECKIN_REDIS_HOST",
		"redis.port":                                 "WECHECKIN_REDIS_PORT",
		"redis.password":                             "WECHECKIN_REDIS_PASSWORD",
		"redis.db":                                   "WECHECKIN_REDIS_DB",
		"log.dir":                                    "WECHECKIN_LOG_DIR",
		"log.level":                                  "WECHECKIN_LOG_LEVEL",
		"log.max_age":                                "WECHECKIN_LOG_MAX_AGE",
		"log.compress":                               "WECHECKIN_LOG_COMPRESS",
		"oss.type":                                   "WECHECKIN_OSS_TYPE",
		"oss.aliyun.access_key_id":                   "WECHECKIN_OSS_ALIYUN_ACCESS_KEY_ID",
		"oss.aliyun.access_key_secret":               "WECHECKIN_OSS_ALIYUN_ACCESS_KEY_SECRET",
		"oss.aliyun.endpoint":                        "WECHECKIN_OSS_ALIYUN_ENDPOINT",
		"oss.aliyun.bucket":                          "WECHECKIN_OSS_ALIYUN_BUCKET",
		"oss.tencent.secret_id":                      "WECHECKIN_OSS_TENCENT_SECRET_ID",
		"oss.tencent.secret_key":                     "WECHECKIN_OSS_TENCENT_SECRET_KEY",
		"oss.tencent.region":                         "WECHECKIN_OSS_TENCENT_REGION",
		"oss.tencent.bucket":                         "WECHECKIN_OSS_TENCENT_BUCKET",
		"oss.local.path":                             "WECHECKIN_OSS_LOCAL_PATH",
		"token.user.expire":                          "WECHECKIN_TOKEN_USER_EXPIRE",
		"token.user.redis_prefix":                    "WECHECKIN_TOKEN_USER_REDIS_PREFIX",
		"token.admin.expire":                         "WECHECKIN_TOKEN_ADMIN_EXPIRE",
		"token.admin.redis_prefix":                   "WECHECKIN_TOKEN_ADMIN_REDIS_PREFIX",
		"cors.allow_origins":                         "WECHECKIN_CORS_ALLOW_ORIGINS",
		"cors.allow_methods":                         "WECHECKIN_CORS_ALLOW_METHODS",
		"cors.allow_headers":                         "WECHECKIN_CORS_ALLOW_HEADERS",
		"scheduled_task.redis_key_prefix":            "WECHECKIN_SCHEDULED_TASK_REDIS_KEY_PREFIX",
		"scheduled_task.worker_count":                "WECHECKIN_SCHEDULED_TASK_WORKER_COUNT",
		"scheduled_task.scheduler_poll_seconds":      "WECHECKIN_SCHEDULED_TASK_SCHEDULER_POLL_SECONDS",
		"scheduled_task.scheduler_recovery_seconds":  "WECHECKIN_SCHEDULED_TASK_SCHEDULER_RECOVERY_SECONDS",
		"scheduled_task.worker_poll_block_seconds":   "WECHECKIN_SCHEDULED_TASK_WORKER_POLL_BLOCK_SECONDS",
		"scheduled_task.worker_heartbeat_seconds":    "WECHECKIN_SCHEDULED_TASK_WORKER_HEARTBEAT_SECONDS",
		"scheduled_task.worker_ttl_seconds":          "WECHECKIN_SCHEDULED_TASK_WORKER_TTL_SECONDS",
		"scheduled_task.recovery_timeout_seconds":    "WECHECKIN_SCHEDULED_TASK_RECOVERY_TIMEOUT_SECONDS",
		"scheduled_task.minimum_second_interval":     "WECHECKIN_SCHEDULED_TASK_MINIMUM_SECOND_INTERVAL",
		"scheduled_task.run_retention_days":          "WECHECKIN_SCHEDULED_TASK_RUN_RETENTION_DAYS",
		"scheduled_task.log_retention_days":          "WECHECKIN_SCHEDULED_TASK_LOG_RETENTION_DAYS",
		"scheduled_task.max_log_segment_bytes":       "WECHECKIN_SCHEDULED_TASK_MAX_LOG_SEGMENT_BYTES",
		"scheduled_task.max_log_run_bytes":           "WECHECKIN_SCHEDULED_TASK_MAX_LOG_RUN_BYTES",
		"scheduled_task.enable_shell":                "WECHECKIN_SCHEDULED_TASK_ENABLE_SHELL",
		"scheduled_task.enable_sql":                  "WECHECKIN_SCHEDULED_TASK_ENABLE_SQL",
		"scheduled_task.http.allowed_hosts":          "WECHECKIN_SCHEDULED_TASK_HTTP_ALLOWED_HOSTS",
		"scheduled_task.http.allowed_cidrs":          "WECHECKIN_SCHEDULED_TASK_HTTP_ALLOWED_CIDRS",
		"scheduled_task.http.allow_private_networks": "WECHECKIN_SCHEDULED_TASK_HTTP_ALLOW_PRIVATE_NETWORKS",
		"scheduled_task.http.max_redirects":          "WECHECKIN_SCHEDULED_TASK_HTTP_MAX_REDIRECTS",
		"scheduled_task.http.max_request_bytes":      "WECHECKIN_SCHEDULED_TASK_HTTP_MAX_REQUEST_BYTES",
		"scheduled_task.http.max_response_bytes":     "WECHECKIN_SCHEDULED_TASK_HTTP_MAX_RESPONSE_BYTES",
	}
	for key, env := range bindings {
		if err := v.BindEnv(key, env); err != nil {
			log.Printf("Warning: bind env %s: %v", env, err)
		}
	}

	for key, env := range map[string]string{
		"cors.allow_origins":                "WECHECKIN_CORS_ALLOW_ORIGINS",
		"cors.allow_methods":                "WECHECKIN_CORS_ALLOW_METHODS",
		"cors.allow_headers":                "WECHECKIN_CORS_ALLOW_HEADERS",
		"scheduled_task.http.allowed_hosts": "WECHECKIN_SCHEDULED_TASK_HTTP_ALLOWED_HOSTS",
		"scheduled_task.http.allowed_cidrs": "WECHECKIN_SCHEDULED_TASK_HTTP_ALLOWED_CIDRS",
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
