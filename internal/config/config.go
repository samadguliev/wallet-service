package config

import (
	"os"
	"strconv"
)

type EnvConfig struct {
	AuthToken string
	Postgres  *Postgres
}

var Config *EnvConfig

type Postgres struct {
	Host     string `toml:"host" yaml:"host" dsl:"host"`
	Port     int    `toml:"port" yaml:"port" dsl:"port"`
	Database string `toml:"db" yaml:"db" dsl:"dbname"`
	User     string `toml:"user" yaml:"user" dsl:"user"`
	Password string `toml:"password" yaml:"password" dsl:"password"`
}

func InitFromFile() {
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))

	Config = &EnvConfig{
		AuthToken: os.Getenv("AUTH_TOKEN"),
		Postgres: &Postgres{
			Host:     os.Getenv("DB_HOST"),
			Port:     port,
			Database: os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
	}
}
