package config

import (
	"github.com/whitewater-guide/gorge/core"
)

type LogConfig struct {
	Level  string `desc:"Log level. Leave empty to discard logs"`
	Format string `desc:"Set this to 'json' to output log in json"`
}

type PgConfig struct {
	Host     string `desc:"Postgres host"`
	Password string `desc:"Postgres password [env POSTGRES_PASSWORD]" env:"~POSTGRES_PASSWORD"`
	User     string `desc:"Postgres user"`
	Db       string `desc:"Postgres database"`
}

type RedisConfig struct {
	Host string `desc:"Redis host"`
	Port string `desc:"Redis port"`
}

type Config struct {
	Endpoint    string `desc:"Endpoint path"`
	Port        string `desc:"Port"`
	Cache       string `desc:"Either 'inmemory' or 'redis'"`
	Db          string `desc:"Either 'inmemory' or 'postgres'"`
	DbChunkSize int    `desc:"Measurements will be saved to db in chunks of this size. When set to 0, they will be saved in one chunk, which can cause errors"`
	Debug       bool   `desc:"Enables debug mode, sets log level to debug"`
	Pg          PgConfig
	Redis       RedisConfig
	Log         LogConfig
	HTTP        core.ClientOptions
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
	}
}

func testConfig() *Config {
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
