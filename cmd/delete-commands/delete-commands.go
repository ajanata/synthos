package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/bots/controller"
	"github.com/ajanata/synthos/internal/config"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &log.Logger

	log.Trace().Msg("Loading config")
	c, err := config.Load()
	if err != nil {
		log.Panic().Err(err).Msg("Error loading config")
	}
	zerolog.SetGlobalLevel(c.SynthOS.LogLevel)

	fmt.Print("Delete ALL global application commands? Type 'yes' to continue: ")
	var text string
	_, err = fmt.Scanln(&text)
	if err != nil {
		log.Panic().Err(err).Msg("Error reading input")
	}
	if text != "yes" {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	b := controller.New(c.SynthOS.Controller, nil)
	err = b.DeleteAllCommands()
	if err != nil {
		log.Panic().Err(err).Msg("Error deleting all commands")
	}
}
