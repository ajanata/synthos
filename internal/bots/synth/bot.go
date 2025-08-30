package synth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/database"
)

const commandEdit = "s;edit "

type Bot struct {
	// the user that owns this synth, that we proxy messages for
	userID string
	token  string

	d *discordgo.Session

	commands        []command
	commandHandlers map[string]cmdHandlerFunc
}

func New(synth *database.Synth) *Bot {
	return &Bot{
		userID: synth.DiscordUserID,
		token:  synth.Token,
	}
}

func (b *Bot) Start() error {
	ctx := log.With().Str("user_id", b.userID).Logger().WithContext(context.Background())
	log.Ctx(ctx).Info().Msg("Starting synth")

	err := b.setup(ctx)
	if err != nil {
		return err
	}

	// TODO more handlers
	b.d.AddHandler(b.messageCreate)
	b.d.AddHandler(b.presenceChanged)
	b.d.AddHandler(b.userChanged)

	// TODO intents
	b.d.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessages

	log.Ctx(ctx).Trace().Msg("Adding handlers")
	// b.d.AddHandler(b.commandHandler)

	log.Ctx(ctx).Trace().Msg("Connecting synth")
	err = b.d.Open()
	if err != nil {
		return fmt.Errorf("opening Discord session: %w", err)
	}

	ctx = b.loggerCtx(ctx)
	err = b.registerCommands(ctx)
	if err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}

	return nil
}

// loggerCtx attaches information about this Synth to a logger in the context.Context.
func (b *Bot) loggerCtx(ctx context.Context) context.Context {
	return log.Ctx(ctx).With().
		Str("user_id", b.userID).
		Str("bot_username", b.d.State.User.Username).
		Logger().WithContext(ctx)
}

// setup configures the bare essentials for the Discord client, but does not connect it.
func (b *Bot) setup(ctx context.Context) error {
	log.Ctx(ctx).Trace().Msg("Setting up synth")

	if b.d != nil {
		return errors.New("bot already setup")
	}

	s, err := discordgo.New("Bot " + b.token)
	if err != nil {
		return fmt.Errorf("creating Discord session: %w", err)
	}

	b.d = s

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Ctx(ctx).Info().
			Str("username", fmt.Sprintf("%v#%v", s.State.User.Username, s.State.User.Discriminator)).
			Msg("Synth logged in")
	})

	return nil
}

func (b *Bot) Close() error {
	return b.d.Close()
}

func (b *Bot) userChanged(s *discordgo.Session, u *discordgo.UserUpdate) {
	fmt.Printf("%+v\n", *u)
}

func (b *Bot) presenceChanged(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	// we only care about our synth user
	if p.User.ID != b.userID {
		return
	}
	fmt.Printf("%+v %+v\n", *p, *p.User)

	st := string(p.Status)
	if st == "offline" {
		st = "invisible"
	}

	err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince:  p.Presence.Since,
		Activities: p.Activities,
		AFK:        false,
		Status:     st,
	})
	if err != nil {
		panic(err)
	}
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// if this message isn't from our synth user, ignore it
	if m.Author.ID != b.userID {
		return
	}

	sendNewMessage := true
	var ref *discordgo.MessageReference

	// if this is a ref to a message in this channel
	if m.MessageReference != nil && m.MessageReference.Type == discordgo.MessageReferenceTypeDefault && m.MessageReference.GuildID == m.GuildID && m.MessageReference.ChannelID == m.ChannelID {
		ref = m.MessageReference

		var err error
		// get the referenced message
		reply, err := s.ChannelMessage(m.MessageReference.ChannelID, m.MessageReference.MessageID)
		if err != nil {
			panic(err)
		}

		// if the ref is a message we sent
		if reply.Author.ID == s.State.User.ID && strings.HasPrefix(m.Content, commandEdit) {
			_, err := s.ChannelMessageEdit(m.ChannelID, reply.ID, m.Content[len(commandEdit):])
			if err != nil {
				panic(err)
			}
			sendNewMessage = false
		}
	}

	if sendNewMessage {
		var flags discordgo.MessageFlags

		// this doesn't seem to be working
		// if m.Flags&discordgo.MessageFlagsSuppressNotifications == discordgo.MessageFlagsSuppressNotifications {
		// 	flags |= discordgo.MessageFlagsSuppressNotifications
		// }

		// if strings.HasPrefix(m.Content, commandNoNotification) {
		// 	// this isn't quite what I want either...
		// 	m.Content = m.Content[len(commandNoNotification):]
		// 	flags |= discordgo.MessageFlagsSuppressNotifications
		// }

		_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:   m.Content + " beep",
			Reference: ref,
			Flags:     flags,
		})
		if err != nil {
			panic(err)
		}
	}

	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		panic(err)
	}
}
