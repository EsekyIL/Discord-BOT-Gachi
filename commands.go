package main

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var ChannelsID []string

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
}
func Commands(s *discordgo.Session, ic *discordgo.InteractionCreate, database *sql.DB) {
	_, lang := SelectDB("channel_log_voiceID", ic.GuildID, database)
	trs := getTranslation(lang)

	if ic.Type == discordgo.InteractionApplicationCommand {
		switch ic.ApplicationCommandData().Name {
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
											Emoji: discordgo.ComponentEmoji{
												Name: "📝",
											},
											// You can also make it a default option, but in this case we won't.
											Default:     false,
											Description: trs.EventLogging,
										},
										{
											Label: trs.Lang,
											Value: "Language_Insert",
											Emoji: discordgo.ComponentEmoji{
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
										Emoji: discordgo.ComponentEmoji{
											Name: "🇺🇦",
										},
									},
									discordgo.Button{
										Label:    "EU",
										Style:    discordgo.SuccessButton,
										Disabled: false,
										CustomID: "Language_EU",
										Emoji: discordgo.ComponentEmoji{
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
			query := fmt.Sprintf(`UPDATE %s SET Language = ? WHERE id = ?`, shortenNumber(ic.GuildID))

			// Використання параметризованих запитів
			statement, err := database.Prepare(query)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer statement.Close()

			// Виконання запиту
			_, err = statement.Exec("EU", ic.GuildID)
			if err != nil {
				fmt.Println("Error executing query:", err)
			}
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
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}
			return

		case "Language_UA":
			query := fmt.Sprintf(`UPDATE %s SET Language = ? WHERE id = ?`, shortenNumber(ic.GuildID))

			// Використання параметризованих запитів
			statement, err := database.Prepare(query)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer statement.Close()

			// Виконання запиту
			_, err = statement.Exec("UA", ic.GuildID)
			if err != nil {
				fmt.Println("Error executing query:", err)
			}
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
			err = s.InteractionRespond(ic.Interaction, response)
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
										Emoji: discordgo.ComponentEmoji{
											Name: "✔️",
										},
									},
									discordgo.Button{
										Label:    trs.Message,
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_message",
										Emoji: discordgo.ComponentEmoji{
											Name: "💬",
										},
									},
									discordgo.Button{
										Label:    trs.VoiceChannels,
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_voice",
										Emoji: discordgo.ComponentEmoji{
											Name: "🎙️",
										},
									},
									discordgo.Button{
										Label:    trs.Events,
										Style:    discordgo.SecondaryButton,
										Disabled: false,
										CustomID: "fd_event",
										Emoji: discordgo.ComponentEmoji{
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
				query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?`, shortenNumber(ic.GuildID))

				// Використання параметризованих запитів
				statement, err := database.Prepare(query)
				if err != nil {
					fmt.Println("Error preparing statement:", err)
					return
				}
				defer statement.Close()

				// Виконання запиту
				_, err = statement.Exec(ChannelsID[0], ChannelsID[1], ChannelsID[2], ic.GuildID)
				if err != nil {
					fmt.Println("Error executing query:", err)
				}

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
			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?`, shortenNumber(ic.GuildID))

			// Використання параметризованих запитів
			statement, err := database.Prepare(query)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer statement.Close()

			// Виконання запиту
			_, err = statement.Exec(ChannelsID[0], ChannelsID[0], ChannelsID[0], ic.GuildID)
			if err != nil {
				fmt.Println("Error executing query:", err)
			}
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

			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?`, shortenNumber(ic.GuildID))

			// Використання параметризованих запитів
			statement, err := database.Prepare(query)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer statement.Close()

			// Виконання запиту
			_, err = statement.Exec(ChannelsID[0], "0", "0", ic.GuildID)
			if err != nil {
				fmt.Println("Error executing query:", err)
			}
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

			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?`, shortenNumber(ic.GuildID))

			// Використання параметризованих запитів
			statement, err := database.Prepare(query)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer statement.Close()

			// Виконання запиту
			_, err = statement.Exec("0", ChannelsID[0], "0", ic.GuildID)
			if err != nil {
				fmt.Println("Error executing query:", err)
			}
			return
		case "fd_event":
			query := fmt.Sprintf(`UPDATE %s SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?`, shortenNumber(ic.GuildID))

			// Використання параметризованих запитів
			statement, err := database.Prepare(query)
			if err != nil {
				fmt.Println("Error preparing statement:", err)
				return
			}
			defer statement.Close()

			// Виконання запиту
			_, err = statement.Exec("0", "0", ChannelsID[0], ic.GuildID)
			if err != nil {
				fmt.Println("Error executing query:", err)
			}
			return
		}
	}
}
