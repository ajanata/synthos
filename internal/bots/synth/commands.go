package synth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/authorizer"
	"github.com/ajanata/synthos/internal/command"
)

func (b *Bot) buildCommands(ctx context.Context) {
	log.Ctx(ctx).Trace().Msg("Building commands")

	b.cmdGroup = command.NewGroup()

	b.cmdGroup.Command("update-avatar").
		Description("Sync your global avatar to your Synth instance").
		Handler(b.updateAvatar).
		InteractionContext(discordgo.InteractionContextBotDM).
		Build()
}

func (b *Bot) updateAvatar(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	ctx = b.loggerCtx(ctx)
	log.Ctx(ctx).Info().Msg("update avatar handler")

	auth := authorizer.Self{}
	authorized, err := auth.Authorized(ctx, b.userID, u)
	if err != nil {
		_ = b.InteractionSimpleTextResponse(s, i.Interaction, "Failed to authorize. SynthOS Controller has been notified.")
		return err
	}
	if !authorized {
		return b.InteractionSimpleTextResponse(s, i.Interaction, "You are not authorized to use this command.")
	}

	// code lifted from discordgo as we want the raw bytes, not an image.Image
	body, err := s.RequestWithBucketID("GET", discordgo.EndpointUserAvatar(u.ID, u.Avatar), nil, discordgo.EndpointUserAvatar("", ""))
	if err != nil {
		_ = b.InteractionSimpleTextResponse(s, i.Interaction, "Failed to download avatar. SynthOS Controller has been notified.")
		return err
	}

	contentType := http.DetectContentType(body)
	b64 := base64.StdEncoding.EncodeToString(body)
	_, err = s.UserUpdate("", fmt.Sprintf("data:%s;base64,%s", contentType, b64), "")
	if err != nil {
		_ = b.InteractionSimpleTextResponse(s, i.Interaction, "Failed to update avatar. SynthOS Controller has been notified.")
		return err
	}

	return b.InteractionSimpleTextResponse(s, i.Interaction, "Avatar updated successfully.")
}
