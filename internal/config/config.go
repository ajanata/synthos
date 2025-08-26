package config

import (
	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"
)

type Config struct {
	meta toml.MetaData

	SynthOS  SynthOS
	Postgres Postgres
}

type SynthOS struct {
	AdminID    string
	LogLevel   zerolog.Level
	Controller ControllerBot
}

type ControllerBot struct {
	Token string
}

type Postgres struct {
	DSN string
}

func Load() (Config, error) {
	var c Config
	meta, err := toml.DecodeFile("synthos.toml", &c)
	c.meta = meta
	return c, err
}
