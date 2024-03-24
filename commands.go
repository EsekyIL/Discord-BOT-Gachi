package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

func ErrorWriter(err error, text string, lineNumber int) {

	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)

	currentTime := time.Now()
	timestamp := currentTime.Format("2006-01-02 15:04:05")

	boldRed.Printf("%s - [%d.line] - %s: %s\n", timestamp, lineNumber, text, err)
}

func registerCommands(sess *discordgo.Session) {

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⛔ Виникла помилка. 🔧 Зверніться у підтримку бота.",
			Flags:   1 << 6,
		},
	}

	cmdMenuLogs := discordgo.ApplicationCommand{
		Name:        "logs",
		Description: "Випадаюче меню з каналами",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        7, // Числове значення для ApplicationCommandOptionTypeChannel
				Name:        "message_id_channel",
				Description: "Введіть ID каналу для логування повідомлень",
				Required:    true,
			},
			{
				Type:        7,
				Name:        "voice_id_channel",
				Description: "Введіть ID каналу для логування голосових каналів",
				Required:    true,
			},
			{
				Type:        7,
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
				Type:        8,
				Name:        "role_id",
				Description: "Введіть ID ролі, яка буде видаватись користувачам",
				Required:    true,
			},
		},
	}
	_, err := sess.ApplicationCommandCreate("1160175895475138611", "", &cmdMenuLogs) // Створення і відправка команд !
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
		if ic.Type == discordgo.InteractionMessageComponent {
			return
		}
		switch {
		//	case ic.ApplicationCommandData().Name == "temp":  Видалення команд
		//		idcmd := ic.ApplicationCommandData().ID
		//		s.ApplicationCommandDelete("1160175895475138611", "", idcmd)
		case ic.ApplicationCommandData().Name == "logs":
			temp := ic.ApplicationCommandData().Options[0].ChannelValue(s)
			channelID_M := temp.ID
			temp = ic.ApplicationCommandData().Options[1].ChannelValue(s)
			channelID_V := temp.ID
			temp = ic.ApplicationCommandData().Options[2].ChannelValue(s)
			channelID_S := temp.ID

			cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				_, _, lineNumber, _ := runtime.Caller(0)
				ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				s.InteractionRespond(ic.Interaction, response)
				return
			}
			section := cfg.Section("LOGS")
			section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue(channelID_M)
			section.Key("CHANNEL_LOGS_VOICE_ID").SetValue(channelID_V)
			section.Key("CHANNEL_LOGS_SERVER_ID").SetValue(channelID_S)
			err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				_, _, lineNumber, _ := runtime.Caller(0)
				ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				s.InteractionRespond(ic.Interaction, response)
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
			role := ic.ApplicationCommandData().Options[2].RoleValue(s, ic.GuildID)

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
			}
			cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				_, _, lineNumber, _ := runtime.Caller(0)
				ErrorWriter(err, "Помилка при завантаженні конфігураційного файлу", lineNumber)
				s.InteractionRespond(ic.Interaction, response)
				return
			}
			section := cfg.Section("EMOJI_REACTIONS")
			section.Key("MESSAGE_REACTION_ID").SetValue(message_ID)
			section.Key("EMOJI_REACTION").SetValue(emoji_string)
			section.Key("ROLE_ADD_ID").SetValue(role.ID)
			err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				_, _, lineNumber, _ := runtime.Caller(0)
				ErrorWriter(err, "Помилка при збереженні конфігураційного файлу", lineNumber)
				s.InteractionRespond(ic.Interaction, response)
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
