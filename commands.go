package main

import (
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

func registerCommands(sess *discordgo.Session) {
	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)

	cmdLogs := &discordgo.ApplicationCommand{ // Створення тіла команди
		Name:        "logs",
		Description: "Налаштування логування на сервері",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message_id_channel",
				Description: "Введіть ID каналу для логування повідомлень",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "voice_id_channel",
				Description: "Введіть ID каналу для логування голосових каналів",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "server_id_channel",
				Description: "Введіть ID каналу для логування серверу (входу, виходу, банів)",
				Required:    true,
			},
		},
	}
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
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "role_id",
				Description: "Введіть ID ролі, яка буде видаватись користувачам",
				Required:    true,
			},
		},
	}
	_, err := sess.ApplicationCommandCreate("1160175895475138611", "", cmdLogs) // Створення і відправка команд
	if err != nil {
		boldRed.Println("Error creating application command,", err)
		return
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", cmdEmojiReactions)
	if err != nil {
		boldRed.Println("Error creating application command,", err)
		return
	}
	sess.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) { // Модуль зчитування команд та збереження результату в файл
		if ic.Type == discordgo.InteractionMessageComponent {
			return
		}
		switch {
		case ic.ApplicationCommandData().Name == "logs":
			channelID_M := ic.ApplicationCommandData().Options[0].StringValue()
			channelID_V := ic.ApplicationCommandData().Options[1].StringValue()
			channelID_S := ic.ApplicationCommandData().Options[2].StringValue()
			switch {
			case len(channelID_M) > 19:
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Довжина першої опції має бути не більше 19 символів",
						Flags:   1 << 6,
					},
				})
				return
			case len(channelID_V) > 19:
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Довжина другої опції має бути не більше 19 символів",
						Flags:   1 << 6,
					},
				})
				return
			case len(channelID_S) > 19:
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Довжина третьої опції має бути не більше 19 символів",
						Flags:   1 << 6,
					},
				})
				return
			}

			cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				_, _, lineNumber, _ := runtime.Caller(0)
				currentTime := time.Now()
				timestamp := currentTime.Format("2006-01-02 15:04:05")
				boldRed.Println(timestamp, " Помилка при завантаженні конфігураційного файлу: [", lineNumber, "] ", err)
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⛔ Виникла помилка. 🔧 Зверніться у підтримку бота.",
						Flags:   1 << 6,
					},
				})
				return
			}
			section := cfg.Section("LOGS")
			section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue(channelID_M)
			section.Key("CHANNEL_LOGS_VOICE_ID").SetValue(channelID_V)
			section.Key("CHANNEL_LOGS_SERVER_ID").SetValue(channelID_S)
			err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				boldRed.Println("Помилка при завантаженні конфігураційного файлу: ", err)
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⛔ Виникла помилка. 🔧 Зверніться у підтримку бота.",
						Flags:   1 << 6,
					},
				})
				return
			}
			s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🎉 Тепер ви можете користуватись логуванням бота! 🎉",
					Flags:   1 << 6,
				},
			})
		case ic.ApplicationCommandData().Name == "reaction":
			message_ID := ic.ApplicationCommandData().Options[0].StringValue()
			emoji_string := ic.ApplicationCommandData().Options[1].StringValue()
			role_ID := ic.ApplicationCommandData().Options[2].StringValue()

			switch {
			case len(message_ID) > 19:
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Довжина першої опції має бути не більше 19 символів",
						Flags:   1 << 6,
					},
				})
				return
			case len(emoji_string) > 10:
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Довжина другої опції має бути не більше 10 символів",
						Flags:   1 << 6,
					},
				})
				return
			case len(role_ID) > 19:
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Довжина третьої опції має бути не більше 19 символів",
						Flags:   1 << 6,
					},
				})
				return
			}
			cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				boldRed.Println("Помилка при завантаженні конфігураційного файлу: ", err)
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⛔ Виникла помилка. 🔧 Зверніться у підтримку бота.",
						Flags:   1 << 6,
					},
				})
				return
			}
			section := cfg.Section("EMOJI_REACTIONS")
			section.Key("MESSAGE_REACTION_ID").SetValue(message_ID)
			section.Key("EMOJI_REACTION").SetValue(emoji_string)
			section.Key("ROLE_ADD_ID").SetValue(role_ID)
			err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				boldRed.Println("Помилка при збереженні конфігураційного файлу: ", err)
				s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⛔ Виникла помилка. 🔧 Зверніться у підтримку бота.",
						Flags:   1 << 6,
					},
				})
				return
			}
			s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🎉 Тепер ви можете користуватись видачею ролей через Emoji! 🎉",
					Flags:   1 << 6,
				},
			})
		}

	})
}
