package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type OptionBuilder struct {
	opt *discordgo.ApplicationCommandOption
	cmd optionable
}

type optionable interface {
	addOption(*discordgo.ApplicationCommandOption)
}

func newOptionBuilder(cmd optionable, name string) *OptionBuilder {
	return &OptionBuilder{
		opt: &discordgo.ApplicationCommandOption{
			Name: name,
		},
		cmd: cmd,
	}
}

func (b *OptionBuilder) Build() {
	if b.opt.Name == "" {
		log.Panic().Msg("Option name is required")
	}
	if b.opt.Description == "" {
		log.Panic().Msg("Option description is required")
	}
	if b.opt.Type == 0 {
		log.Panic().Msg("Option type is required")
	}

	b.cmd.addOption(b.opt)
}

func (b *OptionBuilder) Description(d string) *OptionBuilder {
	b.opt.Description = d
	return b
}

func (b *OptionBuilder) Type(t discordgo.ApplicationCommandOptionType) *OptionBuilder {
	b.opt.Type = t
	return b
}

func (b *OptionBuilder) Required() *OptionBuilder {
	b.opt.Required = true
	return b
}
