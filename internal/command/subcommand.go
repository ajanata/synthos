package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type Subcommand struct {
	subcmd  *discordgo.ApplicationCommandOption
	handler Handler
}

type SubcommandBuilder struct {
	opt     *discordgo.ApplicationCommandOption
	cmd     subcommandable
	handler Handler
}

type subcommandable interface {
	optionable

	addSubcommand(*Subcommand)
}

func (s *Subcommand) addOption(opt *discordgo.ApplicationCommandOption) {
	s.subcmd.Options = append(s.subcmd.Options, opt)
}

func (s *Subcommand) Option(name string) *OptionBuilder {
	return newOptionBuilder(s, name)
}

func newSubcommandBuilder(cmd subcommandable, name string) *SubcommandBuilder {
	return &SubcommandBuilder{
		opt: &discordgo.ApplicationCommandOption{
			Name: name,
			Type: discordgo.ApplicationCommandOptionSubCommand,
		},
		cmd: cmd,
	}
}

func (b *SubcommandBuilder) Build() *Subcommand {
	if b.opt.Name == "" {
		log.Panic().Msg("Subcommand name is required")
	}
	if b.opt.Description == "" {
		log.Panic().Msg("Subcommand description is required")
	}
	if b.opt.Type == 0 {
		log.Panic().Msg("Subcommand type is required")
	}
	if b.handler == nil {
		log.Panic().Msg("Subcommand handler is required")
	}

	s := &Subcommand{
		subcmd:  b.opt,
		handler: b.handler,
	}
	b.cmd.addSubcommand(s)
	return s
}

func (b *SubcommandBuilder) Description(d string) *SubcommandBuilder {
	b.opt.Description = d
	return b
}

func (b *SubcommandBuilder) Handler(handler Handler) *SubcommandBuilder {
	b.handler = handler
	return b
}
