package main

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

func registerCommands(sess *discordgo.Session, database *sql.DB) {

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
	sess.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) { // Модуль зчитування команд та збереження результату в файл
		var response *discordgo.InteractionResponse

		OneValue := ""
		TwoValue := ""
		ThreeValue := ""
		//	case ic.ApplicationCommandData().Name == "temp":  Видалення команд
		//		idcmd := ic.ApplicationCommandData().ID
		//		s.ApplicationCommandDelete("1160175895475138611", "", idcmd)
		embed1 := &discordgo.MessageEmbed{
			Title:       "Налаштування функцій бота",
			Description: "> Виберіть пункт, який хочете налаштувати",
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/o7wcuxw.png",
			},
			Color: 0x6892c2,
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "Kazaki",
				IconURL: "https://i.imgur.com/04X5nxH.png",
			},
		}
		switch ic.Type {
		case discordgo.InteractionApplicationCommand:
			switch ic.ApplicationCommandData().Name {
			case "settings":
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed1},
						Flags:  discordgo.MessageFlagsEphemeral,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.SelectMenu{
										// Select menu, as other components, must have a customID, so we set it to this value.
										CustomID:    "select",
										Placeholder: "Виберіть пункт налаштування 👇",
										Options: []discordgo.SelectMenuOption{
											{
												Label: "Логування",
												// As with components, this things must have their own unique "id" to identify which is which.
												// In this case such id is Value field.
												Value: "logyvanie",
												Emoji: discordgo.ComponentEmoji{
													Name: "📝",
												},
												// You can also make it a default option, but in this case we won't.
												Default:     false,
												Description: "Логування повідомлень, входу/виходу і тд...",
											},
											{
												Label: "Видача ролі по Emoji",
												Value: "js",
												Emoji: discordgo.ComponentEmoji{
													Name: "🎨",
												},
												Description: "Автоматична видача ролі",
											},
											{
												Label: "Python",
												Value: "py",
												Emoji: discordgo.ComponentEmoji{
													Name: "🐍",
												},
												Description: "Python programming language",
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
		case discordgo.InteractionMessageComponent:
			minValues := 1
			if ic.MessageComponentData().CustomID == "select" {
				// Отримання вибору пункту списку
				selectedValue := ic.MessageComponentData().Values[0]
				embed_logs := &discordgo.MessageEmbed{
					Title:       "Налаштування логування серверу",
					Description: ">>> Виберіть канали для логування. Від `одного` до `трьох`. У будь який момент їх можна змінити. ",
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
							Name:  "Вибір першого каналу",
							Value: "*Перший канал виводить вам `зміну/видалення` повідомлень на сервері.*",
						},
						{
							Name:  "Вибір другого каналу",
							Value: "*Другий канал виводить вам `вхід/перехід/вихід` з голосових каналів на сервері.*",
						},
						{
							Name:  "Вибір третього каналу",
							Value: "*Третій канал виводить `вхід/вихід/бан` користувача на сервері.*",
						},
						{
							Name:  "Вибір каналу",
							Value: "***Якщо хочете, щоб логування виводилось в один канал, просто виберіть той канал, який вам потрібен!***",
						},
					},
				}
				// Обробка вибраного значення
				switch selectedValue {
				case "logyvanie":
					response = &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{embed_logs},
							Flags:  discordgo.MessageFlagsEphemeral,
							Components: []discordgo.MessageComponent{
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.SelectMenu{
											MinValues:    &minValues,
											MaxValues:    3,
											MenuType:     discordgo.ChannelSelectMenu,
											CustomID:     "channel_select",
											Placeholder:  "Тута треба тицьнути",
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
				case "fd_yes":
					println("Всі логі бляха муха")
				case "py":
					println("Py its easy")
				}
			}
			if ic.MessageComponentData().CustomID == "channel_select" {
				// Отримання вибору пункту списку
				var Alone bool = false
				var Two bool = false
				var Three bool = false
				switch len(ic.MessageComponentData().Values) {
				case 1:
					OneValue = ic.MessageComponentData().Values[0]
					Alone = true
				case 2:
					OneValue = ic.MessageComponentData().Values[0]
					TwoValue = ic.MessageComponentData().Values[1]
					Two = true
				case 3:
					OneValue = ic.MessageComponentData().Values[0]
					TwoValue = ic.MessageComponentData().Values[1]
					ThreeValue = ic.MessageComponentData().Values[2]
					Three = true
				default:
					TwoValue = OneValue
					ThreeValue = OneValue
				}
				if Alone {
					response = &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:   discordgo.MessageFlagsEphemeral,
							Content: ">>> *Якщо ви хочете всі логи направляти до одного каналу, натисніть кнопку `Усі логи`. Якщо вам потрібне конкретне логування для повідомлень, голосових каналів або подій, виберіть відповідну опцію.*",
							Components: []discordgo.MessageComponent{
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.Button{
											Label:    "Усі логи",
											Style:    discordgo.SuccessButton,
											Disabled: false,
											CustomID: "fd_yes",
											Emoji: discordgo.ComponentEmoji{
												Name: "✔️",
											},
										},
										discordgo.Button{
											Label:    "Повідомлення",
											Style:    discordgo.SecondaryButton,
											Disabled: false,
											CustomID: "fd_message",
											Emoji: discordgo.ComponentEmoji{
												Name: "💬",
											},
										},
										discordgo.Button{
											Label:    "Голосові канали",
											Style:    discordgo.SecondaryButton,
											Disabled: false,
											CustomID: "fd_voice",
											Emoji: discordgo.ComponentEmoji{
												Name: "🎙️",
											},
										},
										discordgo.Button{
											Label:    "Події",
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
				}

				if Two {
					embed_three := &discordgo.MessageEmbed{
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
					response = &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:  discordgo.MessageFlagsEphemeral,
							Embeds: []*discordgo.MessageEmbed{embed_three},
						},
					}
					err = s.InteractionRespond(ic.Interaction, response)
					if err != nil {
						Error("", err)
					}
				}

				if Three {
					statement, _ := database.Prepare("UPDATE servers SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?")
					statement.Exec(OneValue, TwoValue, ThreeValue, ic.GuildID)
					defer statement.Close()

					embed_three := &discordgo.MessageEmbed{
						Title:       "Успішно",
						Color:       0x0ea901,
						Description: "> Тепер можете користуватись логуванням всього серверу",
						Footer: &discordgo.MessageEmbedFooter{
							Text:    "Kazaki",
							IconURL: "https://i.imgur.com/04X5nxH.png",
						},
					}
					response = &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:  discordgo.MessageFlagsEphemeral,
							Embeds: []*discordgo.MessageEmbed{embed_three},
						},
					}
					err = s.InteractionRespond(ic.Interaction, response)
					if err != nil {
						Error("", err)
					}
				}

			}
			if ic.MessageComponentData().CustomID == "fd_yes" {
				statement, _ := database.Prepare("UPDATE servers SET channel_log_msgID = ?, channel_log_voiceID = ?, channel_log_serverID = ? WHERE id = ?")
				statement.Exec(OneValue, OneValue, OneValue, ic.GuildID)
				defer statement.Close()

				embed_AllOne := &discordgo.MessageEmbed{
					Title:       "Успішно",
					Color:       0x0ea901,
					Description: "> Тепер можете користуватись логуванням всього серверу лише в один канал",
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "Kazaki",
						IconURL: "https://i.imgur.com/04X5nxH.png",
					},
				}
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed_AllOne},
					},
				}
				err = s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Помилка при завантаженні конфігураційного файлу", err)
				}
			}
			if ic.MessageComponentData().CustomID == "fd_message" {
				statement, _ := database.Prepare("UPDATE servers SET channel_log_msgID = ? WHERE id = ?")
				statement.Exec(OneValue, ic.GuildID)
				defer statement.Close()

				embed_AllOne := &discordgo.MessageEmbed{
					Title:       "Успішно",
					Color:       0x0ea901,
					Description: "> Тепер можете користуватись тільки логуванням повідомлень",
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "Kazaki",
						IconURL: "https://i.imgur.com/04X5nxH.png",
					},
				}
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed_AllOne},
					},
				}
				err = s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Помилка при завантаженні конфігураційного файлу", err)
				}
			}
			if ic.MessageComponentData().CustomID == "fd_voice" {
				statement, _ := database.Prepare("UPDATE servers SET channel_log_voiceID = ? WHERE id = ?")
				statement.Exec(OneValue, ic.GuildID)
				defer statement.Close()

				embed_AllOne := &discordgo.MessageEmbed{
					Title:       "Успішно",
					Color:       0x0ea901,
					Description: "> Тепер можете користуватись тільки логуванням голосових каналів",
					Footer: &discordgo.MessageEmbedFooter{
						Text:    "Kazaki",
						IconURL: "https://i.imgur.com/04X5nxH.png",
					},
				}
				response = &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:  discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{embed_AllOne},
					},
				}
				err = s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Помилка при завантаженні конфігураційного файлу", err)
				}
			}
			if ic.MessageComponentData().CustomID == "fd_event" {
				statement, _ := database.Prepare("UPDATE servers SET channel_log_serverID = ? WHERE id = ?")
				statement.Exec(OneValue, ic.GuildID)
				defer statement.Close()

			}
			embed_AllOne := &discordgo.MessageEmbed{
				Title:       "Успішно",
				Color:       0x0ea901,
				Description: "> Тепер можете користуватись тільки логуванням входу/виходу/банів на сервер",
				Footer: &discordgo.MessageEmbedFooter{
					Text:    "Kazaki",
					IconURL: "https://i.imgur.com/04X5nxH.png",
				},
			}
			response = &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:  discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{embed_AllOne},
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Помилка при завантаженні конфігураційного файлу", err)
			}
		}
	})
}
