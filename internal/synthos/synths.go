package synthos

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/bots/controller"
	"github.com/ajanata/synthos/internal/bots/synth"
	"github.com/ajanata/synthos/internal/bots/validator"
	"github.com/ajanata/synthos/internal/database"
)

func (app *App) CreateSynth(ctx context.Context, u *discordgo.User, token string) error {
	ctx = log.Ctx(ctx).With().Str("user_id", u.ID).Logger().WithContext(ctx)
	log.Ctx(ctx).Trace().Msg("CreateSynth")

	id, err := validator.GetAppID(ctx, token)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("invalid token")
		return controller.ErrInvalidToken
	}

	return app.db.InsertSynth(ctx, u.ID, id, token)
}

func (app *App) GetSynth(ctx context.Context, u *discordgo.User) (*database.Synth, error) {
	ctx = log.Ctx(ctx).With().Str("user_id", u.ID).Logger().WithContext(ctx)
	log.Ctx(ctx).Trace().Msg("GetSynth")

	return app.db.GetSynth(ctx, u.ID)
}

func (app *App) StartSynth(ctx context.Context, u *discordgo.User) error {
	s, err := app.db.GetSynth(ctx, u.ID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to load synth")
		return controller.ErrUnableToStartSynth
	}

	sb := synth.New(s)
	err = sb.Start()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to start synth")
		return controller.ErrUnableToStartSynth
	}
	app.synths[u.ID] = sb
	return nil
}
