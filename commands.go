package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

var OneValue string
var TwoValue string
var ThreeValue string

func ErrorWriter(err error, text string, lineNumber int) {

	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)

	currentTime := time.Now()
	timestamp := currentTime.Format("2006-01-02 15:04:05")

	boldRed.Printf("%s - [%d.line] - %s: %s\n", timestamp, lineNumber, text, err)
}

func registerCommands(sess *discordgo.Session) {
	cmdEmojiReactions := &discordgo.ApplicationCommand{
		Name:        "reaction",
		Description: "Видача ролі на сервері по emoji",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message_id",
				Description: "Введіть ID повідомлення на якому будуть Emoji",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "Введіть Emoji яке мають натискати користувачі",
				Required:    true,
			},
			{
				Type:        8,
				Name:        "role_id",
				Description: "Введіть ID ролі, яка буде видаватись користувачам",
				Required:    true,
			},
		},
	}

	selectMenu := discordgo.ApplicationCommand{
		Name:        "settings",
		Description: "Choose an option from settings",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err := sess.ApplicationCommandCreate("1160175895475138611", "", &selectMenu) // Створення і відправка команд !
	if err != nil {
		_, _, lineNumber, _ := runtime.Caller(0)
		ErrorWriter(err, "Помилка створення команди програми", lineNumber)
		return
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", cmdEmojiReactions)
	if err != nil {
		_, _, lineNumber, _ := runtime.Caller(0)
		ErrorWriter(err, "Помилка створення команди програми", lineNumber)
		return
	}
	sess.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) { // Модуль зчитування команд та збереження результату в файл
		var response *discordgo.InteractionResponse

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
												Value: "go",
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
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "", lineNumber)
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
				case "go":
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
						_, _, lineNumber, _ := runtime.Caller(0)
						ErrorWriter(err, "", lineNumber)
					}
					println("go language its true ♥")
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
						_, _, lineNumber, _ := runtime.Caller(0)
						ErrorWriter(err, "", lineNumber)
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
						_, _, lineNumber, _ := runtime.Caller(0)
						ErrorWriter(err, "", lineNumber)
					}
				}

				if Three {
					cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
					if err != nil {
						_, _, lineNumber, _ := runtime.Caller(0)
						ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
						return
					}
					section := cfg.Section("LOGS")
					section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue(OneValue)
					section.Key("CHANNEL_LOGS_VOICE_ID").SetValue(TwoValue)
					section.Key("CHANNEL_LOGS_SERVER_ID").SetValue(ThreeValue)
					err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
					if err != nil {
						_, _, lineNumber, _ := runtime.Caller(0)
						ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
						return
					}
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
						_, _, lineNumber, _ := runtime.Caller(0)
						ErrorWriter(err, "", lineNumber)
					}
				}

			}
			if ic.MessageComponentData().CustomID == "fd_yes" {
				cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
				section := cfg.Section("LOGS")
				section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue(OneValue)
				section.Key("CHANNEL_LOGS_VOICE_ID").SetValue(OneValue)
				section.Key("CHANNEL_LOGS_SERVER_ID").SetValue(OneValue)
				err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
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
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				}
			}
			if ic.MessageComponentData().CustomID == "fd_message" {
				cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
				section := cfg.Section("LOGS")
				section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue(OneValue)
				err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
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
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				}
			}
			if ic.MessageComponentData().CustomID == "fd_voice" {
				cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
				section := cfg.Section("LOGS")
				section.Key("CHANNEL_LOGS_VOICE_ID").SetValue(OneValue)
				err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
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
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				}
			}
			if ic.MessageComponentData().CustomID == "fd_event" {
				cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
				}
				section := cfg.Section("LOGS")
				section.Key("CHANNEL_LOGS_SERVER_ID").SetValue(OneValue)
				err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
					return
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
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				}
			}
		}
	})
}
func RoleAddByEmoji(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	cfg, err := ini.Load("servers/" + m.GuildID + "/config.ini")
	if err != nil {
		_, _, lineNumber, _ := runtime.Caller(0)
		ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
		return
	}
	section := cfg.Section("EMOJI_REACTIONS")
	MessageReactionAddID := section.Key("MESSAGE_REACTION_ID").String()
	EmojiReaction := section.Key("EMOJI_REACTION").String()
	addroleID := section.Key("ROLE_ADD_ID").String()
	switch {
	case len(MessageReactionAddID) != 19:
		return
	case len(addroleID) != 19:
		return
	}
	if m.MessageID == MessageReactionAddID {
		// Перевіряємо, чи це потрібна реакція (emoji)
		if m.Emoji.Name == EmojiReaction {
			// Отримуємо ID користувача, який натиснув реакцію
			userID := m.UserID
			member, err := s.GuildMember(m.GuildID, userID)
			if err != nil {
				_, _, lineNumber, _ := runtime.Caller(0)
				ErrorWriter(err, "Помилка отримання учасника", lineNumber)
				return
			}

			// Перевіряємо, чи користувач має певну роль
			hasRole := false
			for _, role := range member.Roles {
				if role == addroleID {
					hasRole = true
					break
				}
			}
			if hasRole {
				// Користувач має певну роль, надсилаємо йому приватне повідомлення
				guild, err := s.Guild(m.GuildID)
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка при отриманні інформації про сервер", lineNumber)
					return
				}
				currentTime := time.Now()
				stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
				channel, err := s.UserChannelCreate(userID)
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка створення каналу", lineNumber)
					return
				}
				// Надсилання приватного повідомлення
				embed := &discordgo.MessageEmbed{
					Title: "Помилка",
					Description: fmt.Sprintf(
						">>> Вам вже видана роль! Якщо ролі немає - зверніться до адміністрації серверу: "+"`%s`",
						guild.Name,
					),
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://i.imgur.com/BKYSMoP.png",
					},
					Color:     0xf5b507, // Колір (у форматі HEX)
					Timestamp: stringTime,
				}
				_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка надсилання повідомлення", lineNumber)
					return
				}
			} else {
				err = s.GuildMemberRoleAdd(m.GuildID, userID, addroleID)
				if err != nil {
					_, _, lineNumber, _ := runtime.Caller(0)
					ErrorWriter(err, "Помилка додавання ролі", lineNumber)
					return
				}
			}
		}
	}
}
