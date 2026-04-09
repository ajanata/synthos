package synth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ajanata/synthos/internal/bots"
	"github.com/ajanata/synthos/internal/command"
	"github.com/ajanata/synthos/internal/database"
)

const commandEdit = "s;edit "
const maxProxyFileSize = 50 * 1024 * 1024

type Bot struct {
	bots.Common

	synth *database.Synth

	d *discordgo.Session

	cmdGroup *command.Group

	// XXX hack
	maxEnergy int
	regen     int
}

func New(synth *database.Synth) *Bot {
	return &Bot{
		synth: synth,
	}
}

func (b *Bot) Start() error {
	ctx := log.With().Str("user_id", b.synth.DiscordUserID).Logger().WithContext(context.Background())
	log.Ctx(ctx).Info().Msg("Starting synth")

	err := b.setup(ctx)
	if err != nil {
		return err
	}

	b.buildCommands(ctx)

	log.Ctx(ctx).Trace().Msg("Adding handlers")
	// TODO more handlers
	b.d.AddHandler(b.messageCreate)
	b.d.AddHandler(b.presenceChanged)
	b.d.AddHandler(b.userChanged)
	b.d.AddHandler(b.interactionHandler)
	b.d.AddHandler(b.disconnectHandler)

	b.d.ShouldReconnectOnError = true
	b.d.ShouldRetryOnRateLimit = true

	// TODO intents
	b.d.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildPresences |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessages

	log.Ctx(ctx).Trace().Msg("Connecting synth")
	err = b.d.Open()
	if err != nil {
		return fmt.Errorf("opening Discord session: %w", err)
	}

	ctx = b.loggerCtx(ctx)
	err = b.cmdGroup.Register(ctx, b.d)
	if err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}

	return nil
}

func (b *Bot) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.cmdGroup.Handler(s, i)
	case discordgo.InteractionMessageComponent, discordgo.InteractionModalSubmit:
		b.configInteractionHandler(s, i)
	default:
		b.trace(b.loggerCtx(context.Background())).
			Str("type", i.Type.String()).
			Str("id", i.ID).
			Msg("Received unknown interaction type; ignoring")
	}
}

// loggerCtx attaches information about this Synth to a logger in the context.Context.
func (b *Bot) loggerCtx(ctx context.Context) context.Context {
	return log.Ctx(ctx).With().
		Str("user_id", b.synth.DiscordUserID).
		Str("bot_username", b.d.State.User.Username).
		Logger().WithContext(ctx)
}

// setup configures the bare essentials for the Discord client, but does not connect it.
func (b *Bot) setup(ctx context.Context) error {
	log.Ctx(ctx).Trace().Msg("Setting up synth")

	if b.d != nil {
		return errors.New("bot already setup")
	}

	s, err := discordgo.New("Bot " + b.synth.Token)
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
	if p.User.ID != b.synth.DiscordUserID {
		return
	}
	fmt.Printf("%+v %+v\n", *p, *p.User)

	st := p.Status
	if st == discordgo.StatusOffline {
		st = discordgo.StatusInvisible
	}

	err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince:  p.Presence.Since,
		Activities: p.Activities,
		AFK:        false,
		Status:     string(st),
	})
	// FIXME this shouldn't be here lol
	if err != nil {
		panic(err)
	}
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := b.loggerCtx(context.Background())
	b.trace(ctx).
		Str("m.Author.Username", m.Author.Username).
		Str("m.ChannelID", m.ChannelID).
		Msg("messageCreate")

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// if this message isn't from our synth user, ignore it
	if m.Author.ID != b.synth.DiscordUserID {
		return
	}

	sendNewMessage := true
	deleteOldMessage := true
	var ref *discordgo.MessageReference

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		panic(err)
	}
	if channel.Type == discordgo.ChannelTypeDM {
		// we can't delete messages in DMs
		deleteOldMessage = false
	}

	// if this is a ref to a message in this channel
	if m.MessageReference != nil &&
		m.MessageReference.Type == discordgo.MessageReferenceTypeDefault &&
		m.MessageReference.GuildID == m.GuildID &&
		m.MessageReference.ChannelID == m.ChannelID {

		ref = m.MessageReference

		var err error
		// get the referenced message
		reply, err := s.ChannelMessage(m.MessageReference.ChannelID, m.MessageReference.MessageID)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Error getting referenced message")
			return
		}

		// if the ref is a message we sent
		if reply.Author.ID == s.State.User.ID && strings.HasPrefix(m.Content, commandEdit) {
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel: m.ChannelID,
				ID:      reply.ID,
				Content: new(m.Content[len(commandEdit):]),
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{
						discordgo.AllowedMentionTypeUsers,
						discordgo.AllowedMentionTypeRoles,
					},
				},
			})
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Error editing message")
				return
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

		var stickerIDs []string
		for _, sticker := range m.StickerItems {
			stickerIDs = append(stickerIDs, sticker.ID)
		}

		var files []*discordgo.File
		// see if any are too large before we download them
		for _, attach := range m.Attachments {
			if attach.Size > maxProxyFileSize {
				_, _ = s.ChannelMessageSend(m.ChannelID, "File too large to proxy (max "+fmt.Sprint(maxProxyFileSize/1024)+" KB)")
				return
			}
		}
		for _, attach := range m.Attachments {
			body, err := s.RequestWithBucketID("GET", attach.URL, nil, "TODO") // TODO rate limit bucket
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Error downloading attachment")
				_, _ = s.ChannelMessageSend(m.ChannelID, "Unable to download attachment!")
				return
			}
			files = append(files, &discordgo.File{
				Name:        attach.Filename,
				Reader:      bytes.NewReader(body),
				ContentType: attach.ContentType,
			})
		}

		_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    m.Content,
			Reference:  ref,
			Flags:      flags,
			StickerIDs: stickerIDs,
			Files:      files,
			Poll:       m.Poll,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeUsers,
					discordgo.AllowedMentionTypeRoles,
				},
			},
		})
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Error sending message")
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unable to proxy message!")
			return
		}
	}

	if deleteOldMessage {
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("Error deleting message")
			return
		}
	}
}

func (b *Bot) deferredEphemeralMessage(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("responding to interaction: %w", err)
	}
	return nil
}

func (b *Bot) trace(ctx context.Context) *zerolog.Event {
	if b.synth.AllowLogging {
		return log.Ctx(ctx).Trace()
	}
	return nil
}

func (b *Bot) disconnectHandler(_ *discordgo.Session, _ *discordgo.Disconnect) {
	ctx := b.loggerCtx(context.Background())
	log.Ctx(ctx).Warn().Msg("Disconnected.")
}
