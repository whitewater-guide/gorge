package config

import (
	"os"

	"github.com/whitewater-guide/gorge/core"
)

type LogConfig struct {
	Level  string `desc:"log level. Leave empty to discard logs"`
	Format string `desc:"set this to 'json' to output log in json"`
}

type PgConfig struct {
	Host     string `desc:"postgres host"`
	Password string `desc:"postgres password [env POSTGRES_PASSWORD]" env:"~POSTGRES_PASSWORD"`
	User     string `desc:"postgres user"`
	Db       string `desc:"postgres database"`
}

type RedisConfig struct {
	Host string `desc:"redis host"`
	Port string `desc:"redis port"`
}

type HealthConfig struct {
	Cron      string   `desc:"cron expression for running health notifier"`
	Threshold int      `desc:"hours required to pass since last successful execution to consider job unhealthy"`
	URL       string   `desc:"external endpoint to call with list of unhealthy jobs"`
	Headers   []string `desc:"headers to set on request, in 'Header: Value' format, similar to curl "`
}

type WebhooksConfig struct {
	Health HealthConfig
}

type Config struct {
	Endpoint    string `desc:"endpoint path"`
	Port        string `desc:"port"`
	Cache       string `desc:"either 'inmemory' or 'redis'"`
	Db          string `desc:"either 'inmemory' or 'postgres'"`
	DbChunkSize int    `desc:"measurements will be saved to db in chunks of this size. When set to 0, they will be saved in one chunk, which can cause errors"`
	Debug       bool   `desc:"enables debug mode, sets log level to debug"`
	Pg          PgConfig
	Redis       RedisConfig
	Log         LogConfig
	HTTP        core.ClientOptions
	Hooks       WebhooksConfig
}

func (cfg *Config) ReadFromEnv() {
	if cfg.Pg.Password == "" {
		cfg.Pg.Password = os.Getenv("POSTGRES_PASSWORD")
	}
}

func NewConfig() *Config {
	return &Config{
		Endpoint: "/",
		Port:     "7080",
		Cache:    "redis",
		Db:       "postgres",
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Pg: PgConfig{
			Host:     "db",
			Password: "",
			User:     "postgres",
			Db:       "postgres",
		},
		Redis: RedisConfig{
			Host: "redis",
			Port: "6379",
		},
		HTTP: core.ClientOptions{
			UserAgent: "whitewater.guide robot",
			Timeout:   60,
		},
		Hooks: WebhooksConfig{
			Health: HealthConfig{
				Cron:      "0 0 * * *",
				Threshold: 48,
			},
		},
	}
}

func TestConfig() *Config {
	return &Config{
		Endpoint: "/",
		Port:     "7080",
		Cache:    "inmemory",
		Db:       "inmemory",
		Log: LogConfig{
			Level:  "",
			Format: "",
		},
		HTTP: core.ClientOptions{
			UserAgent: "test.whitewater.guide robot",
			Timeout:   60,
		},
	}
}
