package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/command"
	"github.com/ajanata/synthos/internal/database"
)

func (b *Bot) buildCommands(ctx context.Context) {
	log.Ctx(ctx).Trace().Msg("Building commands")

	b.cmdGroup = command.NewGroup()
	setup := b.cmdGroup.Command("setup").
		Description("Set up a new Synth for your account").
		Handler(b.setupHandler).
		Build()
	setup.Subcommand("start").
		Description("Start creating a Synth instance").
		Handler(b.setupStartHandler).
		Build()
	token := setup.Subcommand("token").
		Description("Set token for new Synth instance").
		Handler(b.setupTokenHandler).
		Build()
	token.Option("token").
		Description("Discord App Token").
		Type(discordgo.ApplicationCommandOptionString).
		Required().
		Build()
	setup.Subcommand("link").
		Description("Get link for server admins to add Synth to a server, and you to add to your account").
		Handler(b.setupLinkHandler).
		Build()
}

const setupStartMessage = `Hi! This will be formatted better later. For now, deal with it. :sunglasses:

1. Go to https://discord.com/developers/applications and click New Application.
2. Give it a name that is meaningful to you. Maybe ` + "`<your drone identifier>'s SynthOS`" + `. Check the box and hit Create.
3. You can set the icon, display name, and profile information now if you wish, or do it later.
4. Click the Installation tab and make sure Install Link is set to Discord Provided Link.
5. Under Default Install Settings, add "bot" to Guild Install, and select the following permissions:
  * Change Nickname
  * Create Polls
  * Create Public Threads
  * Embed Links
  * Manage Messages
  * Manage Nicknames
  * Manage Threads
  * Send Messages
  * Send Messages in Threads
(This list may change in the future, if something seems like it's not working, send this start command again to see if the list has changed and go update it if needed, and get any server admins to update it too, which might require removing the integration and adding it again.)
6. Click Save Changes.
7. Click the Bot tab and set its username to what you'd like. Maybe ` + "`<your drone identifier>`" + `. Make sure Public Bot is on, if you want it added to servers you are not an admin on. Turn on all 3 options under Privileged Gateway Intents. Click Save Changes.
8. Click Reset Token back up nearer the top, and confirm that you want to do it. Copy that token, you'll need it in the next step. You may wish to save it in a secure location, too, as you won't be able to see it again.
9. Run the ` + "`/setup token <token>` command, where `<token>`" + ` is the value you just copied.
`

func (b *Bot) setupStartHandler(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	log.Ctx(ctx).Info().Msg("setup start handler")

	return b.interactionSimpleTextResponse(s, i.Interaction, setupStartMessage)
}

func (b *Bot) setupTokenHandler(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	log.Ctx(ctx).Info().Msg("setup token handler")

	options := i.ApplicationCommandData().Options
	var content string

	// TODO make this better
	err := b.synther.CreateSynth(ctx, u, options[0].Options[0].StringValue())
	if errors.Is(err, database.ErrAlreadyExists) {
		content = "You already have a Synth instance. You must delete it (TODO) before you can make a new one. If you changed the token, TODO (but for now, delete it (TODO) and make a new one)."
		goto out
	} else if errors.Is(err, ErrInvalidToken) {
		content = "The Discord token is invalid."
		goto out
	} else if errors.Is(err, ErrUnableToStartSynth) {
		content = "Your Synth was created, but was unable to be started."
		goto out
	} else if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error creating synth")
		content = "Unknown error when trying to create Synth instance."
		goto out
	}

	err = b.interactionSimpleTextResponse(s, i.Interaction, "Your Synth has been created, it is now booting! TODO follow up message, but for now, Run `/setup server-link` next.")
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("error sending response")
	}

	err = b.synther.StartSynth(ctx, u)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("error starting synth")
		content = "An internal error occurred while booting your Synth."
	} else {
		content = "Your Synth has been created! Run `/setup server-link` next."
	}

out:
	return b.interactionSimpleTextResponse(s, i.Interaction, content)
}

func (b *Bot) setupLinkHandler(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	log.Ctx(ctx).Info().Msg("setup link handler")

	var content string
	link, err := b.getServerLink(ctx, u)
	if errors.Is(err, database.ErrNotFound) {
		content = "You do not have a Synth instance."
	} else if err != nil {
		log.Ctx(ctx).Err(err).Msg("error getting synth")
		content = "Unknown error when trying to get Synth instance."
	} else {
		content = "Give this link to an admin of each server you'd like your Synth to join: " + link + "\n\nYou should also Add to My Apps."
	}

	return b.interactionSimpleTextResponse(s, i.Interaction, content)
}

func (b *Bot) setupHandler(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	log.Ctx(ctx).Warn().Msg("setup handler called")
	return b.interactionSimpleTextResponse(s, i.Interaction, "This shouldn't be reachable")
}

func (b *Bot) getServerLink(ctx context.Context, u *discordgo.User) (string, error) {
	log.Ctx(ctx).Trace().Msg("getting server link")

	synth, err := b.synther.GetSynth(ctx, u)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s", synth.ApplicationID), nil
}
