package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/config"
	"github.com/ajanata/synthos/internal/database"
	"github.com/ajanata/synthos/internal/synthos"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &log.Logger

	log.Logger.Trace().Msg("Loading config")
	c, err := config.Load()
	if err != nil {
		log.Logger.Panic().Err(err).Msg("Error loading config")
	}
	zerolog.SetGlobalLevel(c.SynthOS.LogLevel)

	log.Logger.Trace().Msg("Connecting to database")
	db, err := database.New(c.Database)
	if err != nil {
		log.Logger.Panic().Err(err).Msg("Error connecting to database")
	}
	log.Logger.Info().Msg("Database connected")

	bot := synthos.New(c, db)

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		bot.Close()
	}()

	err = bot.Run()
	if err != nil {
		log.Logger.Panic().Err(err).Msg("Error starting bot")
	}
}
