package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/database"
)

type cmdHandlerFunc = func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate)
type command struct {
	ac      *discordgo.ApplicationCommand
	handler cmdHandlerFunc
}

func (b *Bot) registerCommands() error {
	log.Trace().Msg("Registering commands")

	b.commands = []command{
		{
			ac: &discordgo.ApplicationCommand{
				Name: "setup",
				// Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextBotDM},
				// IntegrationTypes: &[]discordgo.ApplicationIntegrationType{discordgo.ApplicationIntegrationUserInstall},
				Description: "Set up a new Synth for your account",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type: discordgo.ApplicationCommandOptionSubCommand,
						// ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeDM},
						Name:        "start",
						Description: "Start creating a Synth instance",
					},
					{
						Type: discordgo.ApplicationCommandOptionSubCommand,
						// ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeDM},
						Name:        "token",
						Description: "Set token for new Synth instance",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "token",
								Description: "Discord App token",
								Required:    true,
							},
						},
					},
					{
						Type: discordgo.ApplicationCommandOptionSubCommand,
						// ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeDM},
						Name:        "server-link",
						Description: "Get link for server admins to add Synth to a server",
					},
				},
			},
			handler: b.setupHandler,
		},
	}
	b.commandHandlers = make(map[string]cmdHandlerFunc)

	var appCmds []*discordgo.ApplicationCommand
	for _, c := range b.commands {
		appCmds = append(appCmds, c.ac)
		b.commandHandlers[c.ac.Name] = c.handler
	}

	_, err := b.d.ApplicationCommandBulkOverwrite(b.d.State.User.ID, "", appCmds)
	if err != nil {
		return fmt.Errorf("adding commands: %w", err)
	}

	return nil
}

func (b *Bot) commandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		log.Warn().Str("type", i.Type.String()).Msg("Received incorrect interaction type for a command")
		return
	}

	// log.Trace().Interface("user", i.User).Msg("commandHandler")

	name := i.ApplicationCommandData().Name
	if h, ok := b.commandHandlers[name]; ok {
		ctx := log.With().
			Str("command", name).
			Str("user_id", i.User.ID).
			Str("username", i.User.Username).
			Logger().WithContext(context.Background())
		h(ctx, s, i)
	}
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

func (b *Bot) setupHandler(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Ctx(ctx).Info().Msg("setup handler")

	options := i.ApplicationCommandData().Options

	if len(options) == 0 {
		log.Ctx(ctx).Error().Msg("No options provided")
		_ = b.interactionSimpleTextResponse(s, i.Interaction, "An internal error occurred.")
		return
	}

	var content string
	switch options[0].Name {
	case "start":
		content = setupStartMessage
	case "token":
		// TODO make this better
		err := b.synther.CreateSynth(ctx, i.User, options[0].Options[0].StringValue())
		if errors.Is(err, database.ErrAlreadyExists) {
			content = "You already have a Synth instance. You must delete it (TODO) before you can make a new one. If you changed the token, TODO (but for now, delete it (TODO) and make a new one)."
			break
		} else if errors.Is(err, ErrInvalidToken) {
			content = "The Discord token is invalid."
			break
		} else if errors.Is(err, ErrUnableToStartSynth) {
			content = "Your Synth was created, but was unable to be started."
			break
		} else if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error creating synth")
			content = "Unknown error when trying to create Synth instance."
			break
		}

		err = b.interactionSimpleTextResponse(s, i.Interaction, "Your Synth has been created, it is now booting! TODO follow up message, but for now,  Run `/setup server-link` next.")
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("error sending response")
		}

		err = b.synther.StartSynth(ctx, i.User)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("error starting synth")
			content = "An internal error occurred while booting your Synth."
		} else {
			content = "Your Synth has been created! Run `/setup server-link` next."
		}
	case "server-link":
		link, err := b.getServerLink(ctx, i.User)
		if errors.Is(err, database.ErrNotFound) {
			content = "You do not have a Synth instance."
		} else if err != nil {
			log.Ctx(ctx).Err(err).Msg("error getting synth")
			content = "Unknown error when trying to get Synth instance."
		} else {
			content = "Give this link to an admin of each server you'd like your Synth to join: " + link
		}
	default:
		log.Warn().Str("name", options[0].Name).Msg("Received incorrect subcommand name for setupHandler")
		content = "Uhh, this shouldn't happen."
	}

	err := b.interactionSimpleTextResponse(s, i.Interaction, content)
	if err != nil {
		log.Err(err).Msg("Failed to respond to user in setupHandler")
	}
}

func (b *Bot) getServerLink(ctx context.Context, u *discordgo.User) (string, error) {
	log.Ctx(ctx).Trace().Msg("getting server link")

	synth, err := b.synther.GetSynth(ctx, u)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s", synth.ApplicationID), nil
}
