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
		Description: "select an option in settings",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err := sess.ApplicationCommandCreate("1160175895475138611", "", &selectMenu)
	if err != nil {
		Error("Error creating command settings", err)
		return
	}
	giveawayCreate := discordgo.ApplicationCommand{
		Name:        "gcreate",
		Description: "start giveaway (modal window)",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", &giveawayCreate)
	if err != nil {
		Error("Error creating command gcreate", err)
		return
	}

	ticketCreate := discordgo.ApplicationCommand{
		Name:        "ticket",
		Description: "create ticket in forum",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", &ticketCreate)
	if err != nil {
		Error("Error creating command ticket", err)
		return
	}
	/*err = sess.ApplicationCommandDelete("1160175895475138611", "", "1177394341371707423") ITS FIND AND DELETE COMMANDS, do not erase!!!!!!!!!!
	if err != nil {
		Error("Error delete command", err)
		return
	}
	err = sess.ApplicationCommandDelete("1160175895475138611", "", "1177398356906082368")
	if err != nil {
		Error("Error delete command", err)
		return
	}
	comands, err := sess.ApplicationCommands("1160175895475138611", "")
	if err != nil {
		Error("Error parsing commands", err)
		return
	}
	for _, comand := range comands {
		println(comand.Name)
		println(comand.ID)
		println("=======================")

	}*/
}
func Commands(s *discordgo.Session, ic *discordgo.InteractionCreate) {

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
		case "ticket":
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
									Placeholder: "Ex: 20d 35h 20m",
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
				Title:       "Setting up bot functions",
				Description: "> Select the item you want to adjust",
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
									Placeholder: "Select the setting item 👇",
									Options: []discordgo.SelectMenuOption{
										{
											Label: "Logging in",
											// As with components, this things must have their own unique "id" to identify which is which.
											// In this case such id is Value field.
											Value: "logging",
											Emoji: &discordgo.ComponentEmoji{
												Name: "📝",
											},
											// You can also make it a default option, but in this case we won't.
											Default:     false,
											Description: "Event logging on the server",
										},
										{
											Label: "Ticket system",
											Value: "ticket-sys",
											Emoji: &discordgo.ComponentEmoji{
												Name: "🎟️",
											},
											Description: "Create ticket system on the server",
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
				Error("error send command settings", err)
			}
		}
		return
	} else if ic.Type == discordgo.InteractionMessageComponent {
		switch ic.MessageComponentData().CustomID {
		case "start-ticket":
			var everyoneID string
			roles, err := s.GuildRoles(ic.GuildID)
			if err != nil {
				Error("parsing roles", err)
				return
			}
			for _, role := range roles {
				if role.Name == "@everyone" {
					everyoneID = role.ID
					break
				}
			}
			_, err = s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "tickets",
				Type:     discordgo.ChannelTypeGuildForum,
				Topic:    "The bot automatically creates posts, no intervention is required.",
				Position: 1,
				NSFW:     false,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:    everyoneID,
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
						Deny:  discordgo.PermissionAll,
					},
				},
			})
			if err != nil {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "**Error 😔**" + "\nTo create a forum, you need to make the server a community!",
					},
				}
				err = s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Interaction respond create forum-channel", err)
				}
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "*Success!!! 🙂*",
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond create forum-channel", err)
			}
			return
		case "leave-giveaway":
			_, err := leaveUserGiveaway(ic.GuildID, ic.Interaction.Member.User.ID)
			if err != nil {
				Error("leave user in giveaway", err)
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: ">>> *You have successfully left the giveaway! 😔*",
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond in participates ", err)
			}
			return

		case "participate":
			currentTime := time.Now().Format(time.RFC3339)

			gvw, Participates, err := incrementParticipantCount(ic.GuildID, ic.Interaction.Member.User.ID, ic.Message.ID)
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
										Disabled: false,
										CustomID: "leave-giveaway",
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
				Title:       "Configuring server logging",
				Description: ">>> Select channels to log. From `one` to `three`. You can change them at any time.",
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
						Name:  "Select the first channel",
						Value: "*The first channel shows you `change/delete` messages on the server.*",
					},
					{
						Name:  "Select the second channel",
						Value: "*The second channel gives you `enter/transition/exit` from the voice channels on the server.*",
					},
					{
						Name:  "Selecting the third channel",
						Value: "*The third channel displays `login/logout/ban/kick/timeout` of the user on the server.*",
					},
					{
						Name:  "Channel selection",
						Value: "***If you want the logging output to one channel, just select the channel you need!***",
					},
				},
			}
			switch selectedValue {
			case "ticket-sys":
				embed = &discordgo.MessageEmbed{
					Title: "About ticket system",
					Description: ">>> The ticket system is a feedback system with the `administration/moderation` of the guild.\n" +
						"The bot will automatically create a forum channel. Then the `/ticket` command will be available.\n" +
						"Configure access rights yourself!\n" + "**Recommended: disable the right to create posts for all users.**",
				}
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed},
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "start-ticket",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🚀",
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
			case "logging":
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
										Placeholder:  "It is necessary to poke here",
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
			}
		case "channel_select":
			ChannelsID = ic.MessageComponentData().Values
			switch len(ChannelsID) {
			case 1:
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags: discordgo.MessageFlagsEphemeral,
						Content: ">>> *If you want to send all logs to one channel, click the `All logs` button." +
							"If you need specific logging for messages, voice channels, or events, select the appropriate option.*",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    "All logs",
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "fd_yes",
										Emoji: &discordgo.ComponentEmoji{
											Name: "✔️",
										},
									},
									discordgo.Button{
										Label:    "Message",
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_message",
										Emoji: &discordgo.ComponentEmoji{
											Name: "💬",
										},
									},
									discordgo.Button{
										Label:    "Voice channels",
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_voice",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🎙️",
										},
									},
									discordgo.Button{
										Label:    "Events",
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
					Title:       "CAT",
					Color:       0xfadb84,
					Description: "> I'm lazzy cat xD",
					Image: &discordgo.MessageEmbedImage{
						URL: "https://i.imgur.com/gYaQOEj.jpg",
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "eseky",
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
					Title:       "Successfully",
					Color:       0x0ea901,
					Description: "> You can now use server-wide logging",
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
				Title:       "Successfully",
				Color:       0x0ea901,
				Description: "> Now you can use logging of the whole server in only one channel",
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
				Title:       "Successfully",
				Color:       0x0ea901,
				Description: "> Now you can only use message logging",
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
				Title:       "Successfully",
				Color:       0x0ea901,
				Description: "> Now you can only use voice channel logging",
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
			currentTime := time.Now().Format(time.RFC3339)

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

			go GiveawayCreated(ic.GuildID, prize, description, unixTime, CountParticipate, int(time.Now().Unix()), winners, ic.ChannelID, ic.Interaction.Member.User.ID)

		}
		return
	}
}
