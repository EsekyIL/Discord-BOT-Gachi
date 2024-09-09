package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ChannelsID []string

func ParseFlexibleDuration(input string) (int64, error) {
	// Регулярний вираз для парсингу часу у форматі 20m, 5h, 30s, 4d 20h 25m 10s і подібних
	re := regexp.MustCompile(`(\d+)([smhd])`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format")
	}

	var totalDuration time.Duration

	// Проходимо по всіх знайдених частинах тривалості (наприклад, 4d, 20h, 25m, 10s)
	for _, match := range matches {
		// Парсимо число
		amount, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %v", err)
		}

		// Визначаємо одиницю виміру часу (s = seconds, m = minutes, h = hours, d = days)
		unit := match[2]
		switch unit {
		case "s":
			totalDuration += time.Duration(amount) * time.Second
		case "m":
			totalDuration += time.Duration(amount) * time.Minute
		case "h":
			totalDuration += time.Duration(amount) * time.Hour
		case "d":
			totalDuration += time.Duration(amount) * time.Hour * 24 // Додаємо дні
		default:
			return 0, fmt.Errorf("unknown time unit: %s", unit)
		}
	}

	// Додаємо загальну тривалість до поточного часу і отримуємо UNIX timestamp
	unixTime := time.Now().Add(totalDuration).Unix()
	return unixTime, nil
}
func isAdmin(s *discordgo.Session, ic *discordgo.InteractionCreate) bool {
	member := ic.Interaction.Member.Roles

	for _, roleID := range member {
		role, err := s.State.Role(ic.Interaction.GuildID, roleID)
		if err != nil {
			Error("Error fetching role:", err)
			continue
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}
	return false
}
func registerCommands(sess *discordgo.Session) {

	selectMenu := discordgo.ApplicationCommand{
		Name:        "settings",
		Description: "Виберіть опцію в налаштуваннях",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err := sess.ApplicationCommandCreate("1160175895475138611", "", &selectMenu) // Створення і відправка команд !
	if err != nil {
		Error("Помилка створення команди програми", err)
		return
	}
	giveawayCreate := discordgo.ApplicationCommand{
		Name:        "gcreate",
		Description: "start giveaway (modal window)",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", &giveawayCreate) // Створення і відправка команд !
	if err != nil {
		Error("Помилка створення команди програми", err)
		return
	}
}
func Commands(s *discordgo.Session, ic *discordgo.InteractionCreate) {

	_, lang := SelectDB("channel_log_voiceID", ic.GuildID)
	trs := getTranslation(lang)

	if ic.Type == discordgo.InteractionApplicationCommand {
		if !isAdmin(s, ic) {
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You do not have permission to use this command.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}
			return
		}

		switch ic.ApplicationCommandData().Name {
		case "gcreate":

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "giveaway-create",
					Title:    "Create a giveaway",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "winners",
									Label:     "number of winners",
									Style:     discordgo.TextInputShort,
									MinLength: 1,
									MaxLength: 3,
									Value:     "1",
									Required:  true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "duration",
									Label:       "duration",
									Style:       discordgo.TextInputShort,
									Placeholder: "Ex: 20 minutes",
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "prize",
									Label:       "prize",
									Placeholder: "Enter a title...",
									Style:       discordgo.TextInputShort,
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "description",
									Label:       "description",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Enter the description of the giveaway...",
									Required:    false,
									MinLength:   0,
									MaxLength:   1000,
								},
							},
						},
					},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Giveaway error creating", err)
			}

			return
		case "settings":
			embed := &discordgo.MessageEmbed{
				Title:       trs.SettingFunction,
				Description: fmt.Sprintf("> %s", trs.SelectItem),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://i.imgur.com/o7wcuxw.png",
				},
				Color: 0x6892c2,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Flags:  discordgo.MessageFlagsEphemeral,
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									// Select menu, as other components, must have a customID, so we set it to this value.
									CustomID:    "select",
									Placeholder: trs.SelectSettingItem,
									Options: []discordgo.SelectMenuOption{
										{
											Label: trs.Logging,
											// As with components, this things must have their own unique "id" to identify which is which.
											// In this case such id is Value field.
											Value: "logyvanie",
											Emoji: &discordgo.ComponentEmoji{
												Name: "📝",
											},
											// You can also make it a default option, but in this case we won't.
											Default:     false,
											Description: trs.EventLogging,
										},
										{
											Label: trs.Lang,
											Value: "Language_Insert",
											Emoji: &discordgo.ComponentEmoji{
												Name: "🗣️",
											},
											Description: trs.ChangeLang,
										},
									},
								},
							},
						},
					},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}
		}
		return
	} else if ic.Type == discordgo.InteractionMessageComponent {
		switch ic.MessageComponentData().CustomID {
		case "participate":
			currentTime := time.Now().Format("2006-01-02T15:04:05.999Z07:00")

			gvw, Participates, err := incrementParticipantCount(ic.GuildID, ic.Interaction.Member.User.ID)
			if err != nil {
				Error("incrementParticipantCount", err)
				return
			}
			if Participates {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: ">>> *You have already entered this giveaway!*",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Style:    discordgo.DangerButton,
										Disabled: true,
										CustomID: "test",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🔚",
										},
									},
								},
							},
						},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Interaction respond in participates ", err)
				}
				return
			}
			embed := &discordgo.MessageEmbed{
				Title: gvw.Title,
				Color: 0xfadb84,
				Description: fmt.Sprintf(gvw.Description+"\n\n"+">>> **Ends: **"+"<t:%d:R>"+"  "+"<t:%d:f>"+"\n"+"** Hosted by: **"+"<@%s>"+"\n"+"**Entries: **"+"`%d`"+"\n"+"**Winners: **"+"`%s`",
					gvw.TimeUnix, gvw.TimeUnix, ic.Interaction.Member.User.ID, gvw.CountParticipate, gvw.Winers,
				),
				Footer: &discordgo.MessageEmbedFooter{
					Text:    ic.Interaction.Member.User.Username,
					IconURL: ic.Interaction.Member.User.AvatarURL("256"),
				},
				Timestamp: currentTime,
			}
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style:    discordgo.PrimaryButton,
							Disabled: false,
							CustomID: "participate",
							Emoji: &discordgo.ComponentEmoji{
								Name: "🎆",
							},
						},
					},
				},
			}
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel:    ic.ChannelID,
				ID:         ic.Message.ID,
				Embed:      embed,
				Components: &components, // Тут передаємо слайс без вказівника
			})
			if err != nil {
				Error("Channel message edit complex", err)
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond", err)
			}
			return
		case "select":
			selectedValue := ic.MessageComponentData().Values[0]
			embed := &discordgo.MessageEmbed{
				Title:       trs.ConfigLogging,
				Description: fmt.Sprintf(">>> %s", trs.ChannelsLog),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://i.imgur.com/BKYSMoP.png",
				},
				Color: 0x6892c2,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  trs.SelectFirstChannel,
						Value: trs.FirstChannelDescrip,
					},
					{
						Name:  trs.SelectSecondChannel,
						Value: trs.SecondChannelDescrip,
					},
					{
						Name:  trs.SelectThirdChannel,
						Value: trs.ThirdChannelDescrip,
					},
					{
						Name:  trs.SelectChannel,
						Value: trs.ChannelDescrip,
					},
				},
			}
			switch selectedValue {
			case "logyvanie":
				minValues := 1
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.SelectMenu{
										MinValues:    &minValues,
										MaxValues:    4,
										MenuType:     discordgo.ChannelSelectMenu,
										CustomID:     "channel_select",
										Placeholder:  trs.Placeholder,
										ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
									},
								},
							},
						},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}

				return
			case "Language_Insert":
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: trs.IfChangeLang,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    "UA",
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "Language_UA",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🇺🇦",
										},
									},
									discordgo.Button{
										Label:    "EU",
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "Language_EU",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🇪🇺",
										},
									},
								},
							},
						},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}

				return
			}
		case "Language_EU":
			query := fmt.Sprintf(`UPDATE %s SET Language = EU WHERE id = %s`, shortenNumber(ic.GuildID), ic.GuildID)
			go UpdateDB(query)

			embed := &discordgo.MessageEmbed{
				Title:       "Language Change!",
				Color:       0x5fc437,
				Description: "> The language was successfully changed",
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}
			return

		case "Language_UA":
			query := fmt.Sprintf(`UPDATE %s SET Language = UA WHERE id = %s`, shortenNumber(ic.GuildID), ic.GuildID)
			go UpdateDB(query)

			embed := &discordgo.MessageEmbed{
				Title:       "Мову змінено!",
				Color:       0x5fc437,
				Description: "> Мову було успішно змінено!",
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}
			return
		case "channel_select":
			ChannelsID = ic.MessageComponentData().Values
			switch len(ChannelsID) {
			case 1:
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: trs.BigDescrip,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    trs.AllLogs,
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "fd_yes",
										Emoji: &discordgo.ComponentEmoji{
											Name: "✔️",
										},
									},
									discordgo.Button{
										Label:    trs.Message,
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_message",
										Emoji: &discordgo.ComponentEmoji{
											Name: "💬",
										},
									},
									discordgo.Button{
										Label:    trs.VoiceChannels,
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_voice",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🎙️",
										},
									},
									discordgo.Button{
										Label:    trs.Events,
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_event",
										Emoji: &discordgo.ComponentEmoji{
											Name: "📢",
										},
									},
								},
							},
						},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}

				return
			case 2:
				embed := &discordgo.MessageEmbed{
					Title:       "Незрозуміло",
					Color:       0xffa100,
					Description: "> Функція пока в розробці",
					Image: &discordgo.MessageEmbedImage{
						URL: "https://i.imgur.com/gYaQOEj.jpg",
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "Kazaki",
						IconURL: "https://i.imgur.com/04X5nxH.png",
					},
				}
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}
				return
			case 3:
				embed := &discordgo.MessageEmbed{
					Title:       trs.Success,
					Color:       0x0ea901,
					Description: trs.UseAllLogs,
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "Kazaki",
						IconURL: "https://i.imgur.com/04X5nxH.png",
					},
				}
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}

				// Формування запиту з параметрами
				query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = %s, channel_log_voiceID = %s, channel_log_serverID = %s WHERE id = %s`, shortenNumber(ic.GuildID), ChannelsID[0], ChannelsID[1], ChannelsID[2], ic.GuildID)
				go UpdateDB(query)

				return
			}
		case "fd_yes":
			embed := &discordgo.MessageEmbed{
				Title:       trs.Success,
				Color:       0x0ea901,
				Description: trs.UseAllLogsFirstChannel,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}

			// Формування запиту з параметрами
			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = %s, channel_log_voiceID = %s, channel_log_serverID = %s WHERE id = %s`, shortenNumber(ic.GuildID), ChannelsID[0], ChannelsID[0], ChannelsID[0], ic.GuildID)
			go UpdateDB(query)

			return
		case "fd_message":

			embed := &discordgo.MessageEmbed{
				Title:       trs.Success,
				Color:       0x0ea901,
				Description: trs.UseMessageLog,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}

			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = %s, channel_log_voiceID = 0, channel_log_serverID = 0 WHERE id = %s`, shortenNumber(ic.GuildID), ChannelsID[0], ic.GuildID)
			go UpdateDB(query)

			return
		case "fd_voice":

			embed := &discordgo.MessageEmbed{
				Title:       trs.Success,
				Color:       0x0ea901,
				Description: trs.UseVoiceLog,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}

			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = 0, channel_log_voiceID = %s, channel_log_serverID = 0 WHERE id = %s`, shortenNumber(ic.GuildID), ChannelsID[0], ic.GuildID)
			go UpdateDB(query)

			return
		case "fd_event":
			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = 0, channel_log_voiceID = 0, channel_log_serverID = %s WHERE id = %s`, shortenNumber(ic.GuildID), ChannelsID[0], ic.GuildID)
			go UpdateDB(query)

			return
		}
	} else if ic.Type == discordgo.InteractionModalSubmit {
		switch ic.ModalSubmitData().CustomID {
		case "giveaway-create":
			currentTime := time.Now().Format("2006-01-02T15:04:05.999Z07:00")

			CountParticipate := 0

			var description string
			description = ""

			winners := ic.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			duration := ic.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			prize := ic.ModalSubmitData().Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description = ic.ModalSubmitData().Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			unixTime, err := ParseFlexibleDuration(duration)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: prize,
				Color: 0xfadb84,
				Description: fmt.Sprintf(description+"\n\n"+">>> **Ends: **"+"<t:%d:R>"+"  "+"<t:%d:f>"+"\n"+"** Hosted by: **"+"<@%s>"+"\n"+"**Entries: **"+"`%d`"+"\n"+"**Winners: **"+"`%s`",
					unixTime, unixTime, ic.Interaction.Member.User.ID, CountParticipate, winners,
				),
				Footer: &discordgo.MessageEmbedFooter{
					Text:    ic.Interaction.Member.User.Username,
					IconURL: ic.Interaction.Member.User.AvatarURL("256"),
				},
				Timestamp: currentTime,
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Style:    discordgo.PrimaryButton,
									Disabled: false,
									CustomID: "participate",
									Emoji: &discordgo.ComponentEmoji{
										Name: "🎆",
									},
								},
							},
						},
					},
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}
			go GiveawayCreated(ic.GuildID, prize, description, unixTime, CountParticipate, int(time.Now().Unix()), winners)

		}
		return
	}
}
