package config

import (
	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"
)

type Config struct {
	meta toml.MetaData

	SynthOS  SynthOS
	Database Database
}

type DBDriver string

const (
	PostgresDBDriver DBDriver = "postgres"
	Sqlite3DBDriver           = "sqlite3"
)

type SynthOS struct {
	AdminID    string
	LogLevel   zerolog.Level
	Controller ControllerBot
}

type ControllerBot struct {
	Token string
}

type Database struct {
	DBDriver DBDriver
	DSN      string
}

func Load() (Config, error) {
	var c Config
	meta, err := toml.DecodeFile("synthos.toml", &c)
	c.meta = meta
	return c, err
}
