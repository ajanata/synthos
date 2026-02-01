package synth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
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
	if data.CustomID != "synth_name_"+i.Interaction.Member.User.ID {
		return fmt.Errorf("invalid modal ID: %s", data.CustomID)
	}

	if len(data.Components) < 1 {
		return fmt.Errorf("no components in modal")
	}

	m, err := s.GuildMember(i.GuildID, s.State.User.ID)
	if err != nil {
		return fmt.Errorf("getting member: %w", err)
	}

	newName := m.Nick
	message := "Unable to parse new name"
	if label, ok := data.Components[0].(*discordgo.Label); ok {
		if input, ok := label.Component.(*discordgo.TextInput); ok {
			newName = input.Value
			err = s.GuildMemberNickname(i.GuildID, "@me", newName)
			message = "Synth name set to " + newName
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Str("new_name", newName).Msg("Error setting new name")
				message = "Unable to set new name: " + err.Error()
			}
		}
	}

	return s.InteractionRespond(i.Interaction, b.configMenu(newName, message))
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
		// TODO figure out how to delete the original response, or edit it after the modal
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "synth_name_" + i.Interaction.Member.User.ID,
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
	}

	menu := b.configMenu(currentName, message)
	menu.Type = discordgo.InteractionResponseUpdateMessage
	return s.InteractionRespond(i.Interaction, menu)
}
