package bots

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Common struct{}

func (*Common) InteractionSimpleTextResponse(s *discordgo.Session, i *discordgo.Interaction, msg string) error {
	// was a channel interaction?
	var flags discordgo.MessageFlags
	if i.Member != nil {
		// so we want to only show it to the user that sent the command
		flags = discordgo.MessageFlagsEphemeral
	}
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   flags,
		},
	})

	if err != nil {
		return fmt.Errorf("interaction response: %w", err)
	}
	return nil
}
