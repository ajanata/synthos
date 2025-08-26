package database

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ajanata/synthos/internal/config"
)

type DB struct {
	g *gorm.DB
}

func New(c config.Postgres) (*DB, error) {
	g, err := gorm.Open(postgres.Open(c.DSN), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	db := &DB{
		g: g,
	}

	err = db.migrate()
	if err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	log.Trace().Msg("Migrating database...")

	var err error
	err = db.g.AutoMigrate(&Synth{})
	return err
}
