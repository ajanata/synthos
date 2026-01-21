package synthos

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/bots/controller"
	"github.com/ajanata/synthos/internal/bots/synth"
	"github.com/ajanata/synthos/internal/config"
	"github.com/ajanata/synthos/internal/database"
)

type App struct {
	config config.Config
	db     *database.DB
	close  chan struct{}

	controller *controller.Bot
	synths     map[string]*synth.Bot
}

func New(c config.Config, db *database.DB) *App {
	return &App{
		config: c,
		db:     db,
		close:  make(chan struct{}),
	}
}

// Run blocks until Close is called, or if an error occurs while starting.
func (app *App) Run() error {
	log.Info().Msg("Starting SynthOS")

	defer app.stop()

	log.Trace().Msg("Starting controller")
	app.controller = controller.New(app.config.SynthOS.Controller, app)
	err := app.controller.Start()
	if err != nil {
		return fmt.Errorf("starting controller: %w", err)
	}
	log.Info().Msg("Controller started")

	log.Trace().Msg("Starting synths")
	app.synths = make(map[string]*synth.Bot)
	synths, err := app.db.GetEnabledSynths(context.Background())
	if err != nil {
		return fmt.Errorf("getting enabled synths: %w", err)
	}
	for _, s := range synths {
		sb := synth.New(s, nil) // TODO
		err := sb.Start()
		if err != nil {
			log.Error().Err(err).Str("user_id", s.DiscordUserID).Msg("starting synth")
			continue
		}
		app.synths[s.DiscordUserID] = sb
	}
	log.Info().Msg("Synths started")

	// TODO startup code

	log.Info().Msg("SynthOS started")
	<-app.close
	// stopping is handled by the defer above
	return nil
}

// stop handles stopping things that were started, so we can clean up if there's an error during startup.
func (app *App) stop() {
	log.Info().Msg("Stopping SynthOS")

	// TODO shutdown code
	// while stopping stuff, we want to stop _everything_ even if we get some errors, so we directly log the errors here
	// instead of returning them to our caller

	if app.synths != nil {
		log.Trace().Msg("Stopping synths")
		for uid, sb := range app.synths {
			err := sb.Close()
			if err != nil {
				log.Error().Err(err).Str("user_id", uid).Msg("closing synth")
			}
		}
		log.Info().Msg("Synths stopped")
	}

	if app.controller != nil {
		log.Trace().Msg("Stopping controller")
		err := app.controller.Close()
		if err != nil {
			log.Err(err).Msg("closing controller")
		}
		log.Info().Msg("Controller stopped")
	}
}

func (app *App) Close() {
	close(app.close)
}
