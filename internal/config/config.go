package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
}

type JWTConfig struct {
	SecretKey  string        `mapstructure:"secret_key"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type RabbitMQConfig struct {
	URL string `mapstructure:"url" validate:"required,url"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetTypeByDefaultValue(true)
	if err := viper.BindEnv("jwt.refresh_ttl", "JWT_REFRESH_TTL"); err != nil {
		fmt.Println("Warning binding JWT_REFRESH_TTL")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			fmt.Println("env not found")
		} else {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	bindEnvVars()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(&cfg)

	if cfg.RabbitMQ.URL == "" {
		return nil, fmt.Errorf("RABBITMQ_URL is required")
	}

	fmt.Printf("Config loaded - Host=%s, DB=%s, Port=%s\n",
		cfg.Postgres.Host, cfg.Postgres.DBName, cfg.Postgres.Port)

	return &cfg, nil
}

func bindEnvVars() {
	viper.BindEnv("postgres.host", "POSTGRES_HOST", "DB_HOST")
	viper.BindEnv("postgres.port", "POSTGRES_PORT", "DB_PORT")
	viper.BindEnv("postgres.user", "POSTGRES_USER", "DB_USER")
	viper.BindEnv("postgres.password", "POSTGRES_PASSWORD", "DB_PASSWORD")
	viper.BindEnv("postgres.dbname", "POSTGRES_DBNAME", "DB_NAME", "POSTGRES_DB")

	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "SERVER_MODE")

	viper.BindEnv("jwt.secret_key", "JWT_SECRET_KEY")
	viper.BindEnv("jwt.access_ttl", "JWT_ACCESS_TTL")
	viper.BindEnv("jwt.refresh_ttl", "JWT_REFRESH_TTL")

	viper.BindEnv("rabbitmq.url", "RABBITMQ_URL")
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}

	if cfg.Postgres.Host == "" {
		cfg.Postgres.Host = "localhost"
	}
	if cfg.Postgres.Port == "" {
		cfg.Postgres.Port = "5432"
	}
	if cfg.Postgres.User == "" {
		cfg.Postgres.User = "postgres"
	}
	if cfg.Postgres.DBName == "" {
		cfg.Postgres.DBName = "auth_db_manager"
	}
	if cfg.Postgres.MaxConns <= 0 {
		cfg.Postgres.MaxConns = 25
	}
	if cfg.Postgres.MinConns <= 0 {
		cfg.Postgres.MinConns = 5
	}
	if cfg.Postgres.MaxConnLifetime == 0 {
		cfg.Postgres.MaxConnLifetime = 1 * time.Hour
	}
	if cfg.Postgres.MaxConnIdleTime == 0 {
		cfg.Postgres.MaxConnIdleTime = 30 * time.Minute
	}
}
