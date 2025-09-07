package command

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	cmd      *discordgo.ApplicationCommand
	subcmds  []*Subcommand
	handler  Handler
	handlers map[string]Handler
}

type Handler func(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error

func (c *Command) cmdHandler(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	// check for subcommand
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		subcmd := options[0].Name
		if h, ok := c.handlers[subcmd]; ok {
			return h(ctx, s, u, i)
		}
	}

	// fall back
	return c.handler(ctx, s, u, i)
}

func (c *Command) addOption(opt *discordgo.ApplicationCommandOption) {
	c.cmd.Options = append(c.cmd.Options, opt)
}

func (c *Command) addSubcommand(s *Subcommand) {
	c.cmd.Options = append(c.cmd.Options, s.subcmd)
	c.subcmds = append(c.subcmds, s)
}

func (c *Command) registerSubcommands() {
	c.handlers = make(map[string]Handler)
	for _, subcmd := range c.subcmds {
		c.handlers[subcmd.subcmd.Name] = subcmd.handler
	}
}

// func (c *Command) Options(_ ...*discordgo.ApplicationCommandOption) *Command {
// 	// convenience for code readability
// 	return c
// }

func (c *Command) Option(name string) *OptionBuilder {
	return newOptionBuilder(c, name)
}

// func (c *Command) Subcommands(_ ...*Subcommand) *Command {
// 	// convenience for code readability
// 	return c
// }

func (c *Command) Subcommand(name string) *SubcommandBuilder {
	return newSubcommandBuilder(c, name)
}
