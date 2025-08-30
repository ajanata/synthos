package command

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// Group is a grouping of Commands for a discordgo.Session. There should be only one Group per Session.
type Group struct {
	commands []*Command
	handlers map[string]Handler
}

func NewGroup() *Group {
	return &Group{}
}

func (g *Group) NewCommand() *Builder {
	return newBuilder(g)
}

func (g *Group) Register(ctx context.Context, s *discordgo.Session) error {
	log.Ctx(ctx).Trace().Msg("Registering commands")
	g.handlers = make(map[string]Handler)

	var appCmds []*discordgo.ApplicationCommand
	for _, c := range g.commands {
		appCmds = append(appCmds, c.cmd)
		g.handlers[c.cmd.Name] = c.cmdHandler
	}

	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", appCmds)
	if err != nil {
		return fmt.Errorf("adding commands: %w", err)
	}

	return nil
}

func (g *Group) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		log.Warn().
			Str("type", i.Type.String()).
			Str("id", i.ID).
			Msg("Received incorrect interaction type for a command")
		return
	}

	name := i.ApplicationCommandData().Name
	if h, ok := g.handlers[name]; ok {
		logger := log.With().
			Str("command", name).
			Logger()

		var u *discordgo.User
		// bot DMs have a User
		if i.User != nil {
			u = i.User
			logger.With().
				Str("user_id", i.User.ID).
				Str("username", i.User.Username).
				Logger()
		}

		// but channel messages have a Member
		if i.Member != nil && i.Member.User != nil {
			u = i.Member.User
			logger.With().
				Str("user_id", i.Member.User.ID).
				Str("username", i.Member.User.Username).
				Logger()
		}

		ctx := logger.WithContext(context.Background())
		h(ctx, s, u, i)
	} else {
		log.Warn().
			Str("name", name).
			Msg("No handler found for command")
	}
}
