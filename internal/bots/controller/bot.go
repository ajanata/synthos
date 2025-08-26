package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/config"
	"github.com/ajanata/synthos/internal/database"
)

type Bot struct {
	token string

	synther SynthCRUD

	d *discordgo.Session

	commands        []command
	commandHandlers map[string]cmdHandlerFunc
}

type SynthCRUD interface {
	CreateSynth(ctx context.Context, u *discordgo.User, token string) error
	GetSynth(ctx context.Context, u *discordgo.User) (*database.Synth, error)
	StartSynth(ctx context.Context, u *discordgo.User) error
}

func New(c config.ControllerBot, synther SynthCRUD) *Bot {
	return &Bot{
		token:   c.Token,
		synther: synther,
	}
}

func (b *Bot) Start() error {
	log.Info().Msg("Starting controller")
	err := b.setup()
	if err != nil {
		return err
	}

	// TODO more handlers
	// TODO intents

	log.Trace().Msg("Adding handlers")
	b.d.AddHandler(b.commandHandler)

	log.Trace().Msg("Connecting controller")
	err = b.d.Open()
	if err != nil {
		return fmt.Errorf("opening Discord session: %w", err)
	}

	err = b.registerCommands()
	if err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}

	return nil
}

// setup configures the bare essentials for the Discord client, but does not connect it.
func (b *Bot) setup() error {
	log.Trace().Msg("Setting up controller")

	if b.d != nil {
		return errors.New("bot already setup")
	}

	s, err := discordgo.New("Bot " + b.token)
	if err != nil {
		return fmt.Errorf("creating Discord session: %w", err)
	}

	b.d = s

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().
			Str("username", fmt.Sprintf("%v#%v", s.State.User.Username, s.State.User.Discriminator)).
			Msg("Controller logged in")
	})

	return nil
}

func (b *Bot) Close() error {
	return b.d.Close()
}

// DeleteAllCommands deletes all globally-registered application commands for this bot.
// It creates the connection, deletes the commands, and then closes the connection.
//
// This is intended to be used in a standalone, single-purpose utility.
func (b *Bot) DeleteAllCommands() error {
	log.Info().Msg("Deleting all global application commands")
	err := b.setup()
	if err != nil {
		return err
	}

	log.Trace().Msg("Connecting controller")
	err = b.d.Open()
	if err != nil {
		return fmt.Errorf("opening Discord session: %w", err)
	}

	registeredCommands, err := b.d.ApplicationCommands(b.d.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("getting registered commands: %w", err)
	}
	log.Info().Int("n", len(registeredCommands)).Msg("Loaded registered commands")

	var errs []error
	for _, v := range registeredCommands {
		log.Info().Str("command", v.Name).Msg("Deleting command")
		err := b.d.ApplicationCommandDelete(b.d.State.User.ID, "", v.ID)
		if err != nil {
			log.Err(err).Str("command", v.Name).Msg("Error deleting command")
			errs = append(errs, fmt.Errorf("deleting command '%v': %w", v.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("deleting registered commands: %v", errs)
	}

	return b.Close()
}

func (b *Bot) interactionSimpleTextResponse(s *discordgo.Session, i *discordgo.Interaction, msg string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}
