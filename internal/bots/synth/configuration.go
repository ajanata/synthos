package synth

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	maxAvatarSize = 10 * 1024 * 1024
)

func numericButtonsRow(idPrefix string, curValue int) discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "-10",
				Style:    discordgo.SecondaryButton,
				CustomID: idPrefix + "_m10",
			},
			discordgo.Button{
				Label:    "-1",
				Style:    discordgo.PrimaryButton,
				CustomID: idPrefix + "_m1",
			},
			discordgo.Button{
				Label:    strconv.Itoa(curValue),
				Style:    discordgo.SuccessButton,
				Disabled: true,
				CustomID: idPrefix + "_disp",
			},
			discordgo.Button{
				Label:    "+1",
				Style:    discordgo.PrimaryButton,
				CustomID: idPrefix + "_p1",
			},
			discordgo.Button{
				Label:    "+10",
				Style:    discordgo.SecondaryButton,
				CustomID: idPrefix + "_p10",
			},
		},
	}
}

func (b *Bot) configMenu(currentName, header string) *discordgo.InteractionResponse {
	menu := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
			Components: []discordgo.MessageComponent{
				discordgo.TextDisplay{
					Content: fmt.Sprintf("Configuration options for %s", currentName),
				},
				discordgo.Section{
					Components: []discordgo.MessageComponent{
						discordgo.TextDisplay{Content: "Synth Name: " + currentName},
					},
					Accessory: discordgo.Button{
						Label:    "Change",
						Style:    discordgo.PrimaryButton,
						CustomID: "synth_name",
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Upload Avatar",
							Style:    discordgo.PrimaryButton,
							CustomID: "avatar",
						},
						discordgo.Button{
							Label:    "Change Bio",
							Style:    discordgo.PrimaryButton,
							CustomID: "bio",
						},
					},
				},
				discordgo.Container{
					Components: []discordgo.MessageComponent{
						discordgo.TextDisplay{
							Content: "Maximum Energy",
						},
						numericButtonsRow("max", b.maxEnergy),
					},
				},
				discordgo.Container{
					Components: []discordgo.MessageComponent{
						discordgo.TextDisplay{
							Content: "Energy Regen per Minute",
						},
						numericButtonsRow("regen", b.regen),
					},
				},
			},
		},
	}

	if header != "" {
		menu.Data.Components = append([]discordgo.MessageComponent{discordgo.TextDisplay{Content: header}}, menu.Data.Components...)
	}

	return menu
}

func (b *Bot) configure(ctx context.Context, s *discordgo.Session, u *discordgo.User, i *discordgo.InteractionCreate) error {
	ctx = b.loggerCtx(ctx)
	log.Ctx(ctx).Info().Msg("configure handler")

	if auth, err := b.authorized(ctx, s, u, i); err != nil || !auth {
		return err
	}

	m, err := s.GuildMember(i.GuildID, s.State.User.ID)
	if err != nil {
		return fmt.Errorf("getting member: %w", err)
	}

	return s.InteractionRespond(i.Interaction, b.configMenu(m.Nick, ""))
}

func (b *Bot) configHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := b.loggerCtx(context.Background())

	var err error
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		err = b.configMessageComponentHandler(ctx, s, i)
	case discordgo.InteractionModalSubmit:
		err = b.configModalHandler(ctx, s, i)
	default:
		if i.Type != discordgo.InteractionMessageComponent && i.Type != discordgo.InteractionModalSubmit {
			log.Ctx(ctx).Trace().
				Str("type", i.Type.String()).
				Str("id", i.ID).
				Msg("Received incorrect interaction type for a message component; ignoring")
		}
	}

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Error handling interaction")
	}
}

func (b *Bot) configModalHandler(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	log.Ctx(ctx).Info().Msg("config modal handler")

	data := i.ModalSubmitData()
	modalID := strings.Split(data.CustomID, "_")
	if len(modalID) != 2 {
		return fmt.Errorf("invalid modal ID (has %d parts, not 2): %s", len(modalID), data.CustomID)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("responding to interaction: %w", err)
	}

	modalType := modalID[0]

	if len(data.Components) < 1 {
		return fmt.Errorf("no components in modal")
	}

	m, err := s.GuildMember(i.GuildID, s.State.User.ID)
	if err != nil {
		return fmt.Errorf("getting member: %w", err)
	}

	name := m.Nick
	var message string

	switch modalType {
	case "synthName":
		message = "Unable to parse new name"
		if label, ok := data.Components[0].(*discordgo.Label); ok {
			if input, ok := label.Component.(*discordgo.TextInput); ok {
				name = input.Value
				err = s.GuildMemberNickname(i.GuildID, "@me", name)
				message = "Synth name set to " + name
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Str("new_name", name).Msg("Error setting new name")
					message = "Unable to set new name: " + err.Error()
				}
			} else {
				return fmt.Errorf("malformed interaction data")
			}
		} else {
			return fmt.Errorf("malformed interaction data")
		}
	case "bio":
		message = "Unable to parse new bio"
		if label, ok := data.Components[1].(*discordgo.Label); ok {
			if input, ok := label.Component.(*discordgo.TextInput); ok {
				bio := input.Value
				_, err = s.GuildCurrentMemberEdit(i.GuildID, &discordgo.GuildCurrentMemberParams{
					Bio: &bio,
				})
				message = "Synth bio updated"
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Str("new_bio", bio).Msg("Error setting new bio")
					message = "Unable to set new bio: " + err.Error()
				}
			} else {
				return fmt.Errorf("malformed interaction data")
			}
		} else {
			return fmt.Errorf("malformed interaction data")
		}
	case "avatar":
		message = "Unable to parse new avatar"

		if label, ok := data.Components[1].(*discordgo.Label); ok {
			if avatar, ok := label.Component.(*discordgo.FileUpload); ok {
				log.Ctx(ctx).Info().Msg("changing avatar")
				_ = avatar
				if len(avatar.Values) != 1 {
					return fmt.Errorf("malformed interaction data")
				}
				att := data.Resolved.Attachments[avatar.Values[0]]
				if att.Size > maxAvatarSize {
					message = fmt.Sprintf("Avatar too large (max %d bytes, was %d bytes)", maxAvatarSize, att.Size)
				} else if att.ContentType != "image/png" && att.ContentType != "image/jpeg" {
					message = fmt.Sprintf("Avatar must be a PNG or JPEG (was %s)", att.ContentType)
				} else {
					body, err := s.RequestWithBucketID("GET", att.URL, nil, "ephemeral-attachments")
					if err != nil {
						return fmt.Errorf("downloading avatar: %w", err)
					}
					b64 := base64.StdEncoding.EncodeToString(body)
					_, err = s.UserUpdate("", fmt.Sprintf("data:%s;base64,%s", att.ContentType, b64), "")
					if err != nil {
						// TODO lots of logging in this func
						message = "Failed to update avatar. TODO SynthOS Controller has been notified."
					}
					message = "Avatar updated successfully."
				}
			} else {
				return fmt.Errorf("malformed interaction data")
			}
		} else {
			return fmt.Errorf("malformed interaction data")
		}
	default:
		return fmt.Errorf("invalid modal response type: %s", modalType)
	}

	menu := b.configMenu(name, message)
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Flags:      discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
		Components: &menu.Data.Components,
	})
	return err
}

func (b *Bot) configMessageComponentHandler(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	log.Ctx(ctx).Info().Msg("config message component handler")

	m, err := s.GuildMember(i.GuildID, s.State.User.ID)
	if err != nil {
		return fmt.Errorf("getting member: %w", err)
	}
	currentName := m.Nick

	message := "Unknown interaction"
	data := i.MessageComponentData()
	switch data.CustomID {
	case "max_m10":
		b.maxEnergy -= 10
		message = "Maximum energy set to " + strconv.Itoa(b.maxEnergy)
	case "max_m1":
		b.maxEnergy -= 1
		message = "Maximum energy set to " + strconv.Itoa(b.maxEnergy)
	case "max_p1":
		b.maxEnergy += 1
		message = "Maximum energy set to " + strconv.Itoa(b.maxEnergy)
	case "max_p10":
		b.maxEnergy += 10
		message = "Maximum energy set to " + strconv.Itoa(b.maxEnergy)
	case "regen_m10":
		b.regen -= 10
		message = "Energy regen set to " + strconv.Itoa(b.regen)
	case "regen_m1":
		b.regen -= 1
		message = "Energy regen set to " + strconv.Itoa(b.regen)
	case "regen_p1":
		b.regen += 1
		message = "Energy regen set to " + strconv.Itoa(b.regen)
	case "regen_p10":
		b.regen += 10
		message = "Energy regen set to " + strconv.Itoa(b.regen)
	case "synth_name":
		// TODO figure out how to delete the original response, or edit it after the modal, if possible
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: fmt.Sprintf("synthName_%s", i.Interaction.ID),
				Title:    "Set Synth Name",
				Flags:    discordgo.MessageFlagsIsComponentsV2,
				Components: []discordgo.MessageComponent{
					discordgo.Label{
						Label:       "Synth Name",
						Description: "Enter a new name for this synth",
						Component: discordgo.TextInput{
							CustomID:  "synth_name",
							Style:     discordgo.TextInputShort,
							MinLength: 1,
							MaxLength: 32,
							Required:  true,
							Value:     m.Nick,
						},
					},
				},
			},
		})
	case "avatar":
		// TODO figure out how to delete the original response, or edit it after the modal, if possible
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: fmt.Sprintf("avatar_%s", i.Interaction.ID),
				Title:    "Set Bio",
				Flags:    discordgo.MessageFlagsIsComponentsV2,
				Components: []discordgo.MessageComponent{
					discordgo.TextDisplay{
						Content: "TODO current avatar?",
					},
					discordgo.Label{
						Label:       "New avatar",
						Description: "Upload a new avatar for this synth",
						Component: discordgo.FileUpload{
							CustomID:  "avatar",
							Required:  new(true),
							MaxValues: 1,
						},
					},
				},
			},
		})
	case "bio":
		// TODO figure out how to delete the original response, or edit it after the modal, if possible
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: fmt.Sprintf("bio_%s", i.Interaction.ID),
				Title:    "Set Bio",
				Flags:    discordgo.MessageFlagsIsComponentsV2,
				Components: []discordgo.MessageComponent{
					discordgo.TextDisplay{
						Content: "It is not currently possible to display the current bio here due to intentional Discord privacy controls.",
					},
					discordgo.Label{
						Label:       "Bio",
						Description: "Enter a new bio for this synth",
						Component: discordgo.TextInput{
							CustomID:  "synth_bio",
							Style:     discordgo.TextInputParagraph,
							MinLength: 1,
							MaxLength: 200,
							Required:  true,
						},
					},
				},
			},
		})
	}

	menu := b.configMenu(currentName, message)
	menu.Type = discordgo.InteractionResponseUpdateMessage
	return s.InteractionRespond(i.Interaction, menu)
}
