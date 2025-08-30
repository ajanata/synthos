package synth

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type cmdHandlerFunc = func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate)
type command struct {
	ac      *discordgo.ApplicationCommand
	handler cmdHandlerFunc
}

func (b *Bot) registerCommands(ctx context.Context) error {
	log.Ctx(ctx).Trace().Msg("Registering commands")

	b.commands = []command{}

	return nil
}
