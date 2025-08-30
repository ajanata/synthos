package command

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	cmd     *discordgo.ApplicationCommand
	handler Handler
}

type Handler func(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate)

func (c *Command) cmdHandler(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) {
	// TODO do we actually need anything at this layer?
	c.handler(ctx, s, u, i)
}
