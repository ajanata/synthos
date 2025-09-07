package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type Builder struct {
	cmd     *discordgo.ApplicationCommand
	handler Handler
	grp     *Group
}

func newBuilder(grp *Group, name string) *Builder {
	return &Builder{
		cmd: &discordgo.ApplicationCommand{
			Name: name,
		},
		grp: grp,
	}
}

func (b *Builder) Build() *Command {
	if b.cmd.Name == "" {
		log.Panic().Msg("Command name is required")
	}
	if b.cmd.Description == "" {
		log.Panic().Msg("Command description is required")
	}
	if b.handler == nil {
		log.Panic().Msg("Command handler is required")
	}

	c := &Command{
		cmd:     b.cmd,
		handler: b.handler,
	}
	b.grp.commands = append(b.grp.commands, c)
	return c
}

func (b *Builder) Description(d string) *Builder {
	b.cmd.Description = d
	return b
}

func (b *Builder) Handler(h Handler) *Builder {
	b.handler = h
	return b
}
