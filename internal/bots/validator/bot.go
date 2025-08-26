package validator

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func GetAppID(ctx context.Context, token string) (string, error) {
	log.Ctx(ctx).Info().Msg("Validating token")

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return "", fmt.Errorf("creating Discord session: %w", err)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Ctx(ctx).Info().
			Str("bot_username", r.User.Username).
			Str("bot_user_id", r.User.ID).
			Msg("Token validated")
	})

	err = s.Open()
	if err != nil {
		return "", fmt.Errorf("opening Discord session: %w", err)
	}

	id := s.State.Application.ID
	_ = s.Close()
	return id, nil
}
