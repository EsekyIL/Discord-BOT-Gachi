package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/ini.v1"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	const Token = "MTE2MDE3NTg5NTQ3NTEzODYxMQ.GfSoxQ.Q1jTHNBHIIQcPAL_pnUmnih8OecIC2sZIp9JvU"
	const commandPrefix = "!"
	var userChannels map[string]string
	userChannels = make(map[string]string)
	var userTimeJoin map[string]string
	userTimeJoin = make(map[string]string)
	var userTimeJoinVoice map[string]string
	userTimeJoinVoice = make(map[string]string)
	sess, err := discordgo.New("Bot " + Token) // Відкриття сессії з ботом
	if err != nil {
		log.Fatal(err)
	}
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
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", cmdLogs) // Створення і відправка команд
	if err != nil {
		fmt.Println("Error creating application command,", err)
		return
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", cmdEmojiReactions)
	if err != nil {
		fmt.Println("Error creating application command,", err)
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

			s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Тепер ви можете користуватись логуванням бота!",
					Flags:   1 << 6,
				},
			})
			cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
				return
			}
			section := cfg.Section("LOGS")
			section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue(channelID_M)
			section.Key("CHANNEL_LOGS_VOICE_ID").SetValue(channelID_V)
			section.Key("CHANNEL_LOGS_SERVER_ID").SetValue(channelID_S)
			err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				fmt.Println("Помилка при збереженні конфігураційного файлу:", err)
				return
			}
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
			s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Тепер ви можете користуватись видачею ролей через Emoji!",
					Flags:   1 << 6,
				},
			})
			cfg, err := ini.Load("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
				return
			}
			section := cfg.Section("EMOJI_REACTIONS")
			section.Key("MESSAGE_REACTION_ID").SetValue(message_ID)
			section.Key("EMOJI_REACTION").SetValue(emoji_string)
			section.Key("ROLE_ADD_ID").SetValue(role_ID)
			err = cfg.SaveTo("servers/" + ic.GuildID + "/config.ini")
			if err != nil {
				fmt.Println("Помилка при збереженні конфігураційного файлу:", err)
				return
			}
		}

	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.GuildCreate) { // Модуль створення папки серверу, конфігураційного файлу а також лога повідомлень
		basePath := "./servers"
		folderName := m.Guild.ID
		folderPath := filepath.Join(basePath, folderName)
		_, err = os.Stat(folderPath)
		if os.IsNotExist(err) {
			err = os.Mkdir(folderPath, 0755)
		} else {
			return
		}
		directoryPath := filepath.Join(basePath, folderName)
		filePath := filepath.Join(directoryPath, "config.ini")
		cfg := ini.Empty()
		section := cfg.Section("GUILD")
		section.Key("GUILD_NAME").SetValue(m.Guild.Name)
		section.Key("GUILD_ID").SetValue(m.Guild.ID)
		section.Key("GUILD_REGION").SetValue(m.Guild.Region)
		section = cfg.Section("LOGS")
		section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue("")
		section.Key("CHANNEL_LOGS_VOICE_ID").SetValue("")
		section.Key("CHANNEL_LOGS_SERVER_ID").SetValue("")
		section = cfg.Section("EMOJI_REACTIONS")
		section.Key("MESSAGE_REACTION_ID").SetValue("")
		section.Key("EMOJI_REACTION").SetValue("")
		section.Key("ROLE_ADD_ID").SetValue("")
		section = cfg.Section("LVL_EXP_USERS")
		members, err := s.GuildMembers(m.Guild.ID, "", 1000)
		if err != nil {
			fmt.Println("Помилка отримання учасників сервера:", err)
			return
		}
		for _, member := range members {
			switch {
			case member.User.ID == "1160175895475138611":
				continue
			case len(member.Roles) == 0:
				continue
			}
			section.Key(member.User.ID).SetValue("0")
		}
		err = cfg.SaveTo(filePath)
		if err != nil {
			fmt.Println("Помилка при збереженні у файл:", err)
			return
		}
		var logger *log.Logger
		l := &lumberjack.Logger{
			Filename:   "servers/" + m.Guild.ID + "/message.log",
			MaxSize:    8192, // мегабайти
			MaxBackups: 1,
			MaxAge:     30, // дні
		}
		logger = log.New(l, "", log.LstdFlags)
		logger.Println("Привіт, цей бот був написаний ручками 𝕙𝕥𝕥𝕡𝕤://𝕥.𝕞𝕖/𝔼𝕤𝕖𝕜𝕪𝕚𝕝 ♥")
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) { // Модуль відстеження повідомлень, а також запис їх у log
		if m.Author.Bot {
			return
		}
		cfg, err := ini.Load("servers/" + m.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsMessages := section.Key("CHANNEL_LOGS_MESSAGE_ID").String()
		ChannelLogsVoice := section.Key("CHANNEL_LOGS_VOICE_ID").String()
		ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
		if m.ChannelID == ChannelLogsMessages || m.ChannelID == ChannelLogsVoice || m.ChannelID == ChannelLogsServer {
			return
		} else {
			filePath := filepath.Join("servers", m.GuildID, "message.log")
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			logger := log.New(file, "", log.LstdFlags)
			logger.Println("Text message: " + m.Content + " | " + "Nickname: " + m.Author.Username + " | " + "ID: " + m.Author.ID + " | " + "messageID: " + m.Message.ID + " | " + "ChannelID: " + m.ChannelID)
			if len(m.Content) >= 3 && len(m.Content) <= 10 {
				var EXP uint32
				section = cfg.Section("LVL_EXP_USERS")
				valueStr := section.Key(m.Author.ID).String()
				parsedEXP, err := strconv.ParseUint(valueStr, 10, 32)
				if err != nil {
					fmt.Println("Помилка конвертації значення рядка в uint32:", err)
					// Обробка помилки, якщо потрібно
					return
				}
				EXP = uint32(parsedEXP)
				EXP += 2
				EXPStr := strconv.Itoa(int(EXP))
				section.Key(m.Author.ID).SetValue(EXPStr)
				basePath := "./servers"
				folderName := m.GuildID
				directoryPath := filepath.Join(basePath, folderName)
				filePath = filepath.Join(directoryPath, "config.ini")
				err = cfg.SaveTo(filePath)
				if err != nil {
					fmt.Println("Помилка при збереженні у файл:", err)
					return
				}
			} else if len(m.Content) > 10 && len(m.Content) < 20 {
				var EXP uint32
				section = cfg.Section("LVL_EXP_USERS")
				valueStr := section.Key(m.Author.ID).String()
				parsedEXP, err := strconv.ParseUint(valueStr, 10, 32)
				if err != nil {
					fmt.Println("Помилка конвертації значення рядка в uint32:", err)
					// Обробка помилки, якщо потрібно
					return
				}
				EXP = uint32(parsedEXP)
				EXP += 5
				EXPStr := strconv.Itoa(int(EXP))
				section.Key(m.Author.ID).SetValue(EXPStr)
				basePath := "./servers"
				folderName := m.GuildID
				directoryPath := filepath.Join(basePath, folderName)
				filePath = filepath.Join(directoryPath, "config.ini")
				err = cfg.SaveTo(filePath)
				if err != nil {
					fmt.Println("Помилка при збереженні у файл:", err)
					return
				}
			} else if len(m.Content) >= 20 {
				var EXP uint32
				section = cfg.Section("LVL_EXP_USERS")
				valueStr := section.Key(m.Author.ID).String()
				parsedEXP, err := strconv.ParseUint(valueStr, 10, 32)
				if err != nil {
					fmt.Println("Помилка конвертації значення рядка в uint32:", err)
					// Обробка помилки, якщо потрібно
					return
				}
				EXP = uint32(parsedEXP)
				EXP += 10
				EXPStr := strconv.Itoa(int(EXP))
				section.Key(m.Author.ID).SetValue(EXPStr)
				basePath := "./servers"
				folderName := m.GuildID
				directoryPath := filepath.Join(basePath, folderName)
				filePath = filepath.Join(directoryPath, "config.ini")
				err = cfg.SaveTo(filePath)
				if err != nil {
					fmt.Println("Помилка при збереженні у файл:", err)
					return
				}
			}
		}
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) { // Модуль додавання ролі по реакції на повідомлення
		cfg, err := ini.Load("servers/" + m.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
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
					fmt.Println("error getting member:", err)
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
					guild, err := sess.Guild(m.GuildID)
					if err != nil {
						fmt.Println("Помилка при отриманні інформації про сервер:", err)
						return
					}
					currentTime := time.Now()
					stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
					channel, err := s.UserChannelCreate(userID)
					if err != nil {
						fmt.Println("error creating channel:", err)
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
						fmt.Println("error sending message:", err)
						return
					}
				} else {
					err = s.GuildMemberRoleAdd(m.GuildID, userID, addroleID)
					if err != nil {
						fmt.Println("error adding role,", err)
						return
					}
					basePath := "./servers"
					folderName := m.GuildID
					directoryPath := filepath.Join(basePath, folderName)
					filePath := filepath.Join(directoryPath, "config.ini")
					section = cfg.Section("LVL_EXP_USERS")
					section.Key(m.UserID).SetValue("0")
					err = cfg.SaveTo(filePath)
					if err != nil {
						fmt.Println("Помилка при збереженні у файл:", err)
						return
					}
				}
			}
		}

	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) { // Модуль логування оновленого повідомлення, а також запис у log
		if m.Author == nil || m.Author.Bot {
			return
		}
		cfg, err := ini.Load("servers/" + m.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsMessages := section.Key("CHANNEL_LOGS_MESSAGE_ID").String()
		ChannelLogsVoice := section.Key("CHANNEL_LOGS_VOICE_ID").String()
		ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
		switch {
		case len(ChannelLogsMessages) != 19:
			return
		case len(ChannelLogsVoice) != 19:
			return
		case len(ChannelLogsServer) != 19:
			return
		}
		if m.ChannelID == ChannelLogsMessages || m.ChannelID == ChannelLogsServer || m.ChannelID == ChannelLogsVoice {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		MessageUpdateID := m.Message.ID
		UserMessage := ""
		file, err := os.OpenFile("servers/"+m.GuildID+"/message.log", os.O_RDWR, 0644)
		if err != nil {
			fmt.Println("Помилка відкриття файлу:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, MessageUpdateID) {
				re := regexp.MustCompile(`Text message: ([^|]+)`)
				match := re.FindStringSubmatch(line)
				if len(match) > 1 {
					UserMessage = match[1]
					_, err := file.Seek(int64(-len(line)), io.SeekCurrent)
					if err != nil {
						fmt.Println("error seeking:", err)
						return
					}
					filePath := filepath.Join("servers", m.GuildID, "message.log")
					file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()
					logger := log.New(file, "", log.LstdFlags)
					logger.Println("Text message: " + m.Content + " | " + "Nickname: " + m.Author.Username + " | " + "ID: " + m.Author.ID + " | " + "messageID: " + m.Message.ID + " | " + "ChannelID: " + m.ChannelID)
					break
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Помилка при читанні файлу:", err)
			return
		}
		embed := &discordgo.MessageEmbed{
			Title: "Повідомлення оновлено",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "**Канал**",
					Value:  "<#" + m.ChannelID + ">",
					Inline: true,
				},
				{
					Name:   "**Автор**",
					Value:  "<@" + m.Author.ID + ">",
					Inline: true,
				},
			},
			Description: fmt.Sprintf(
				">>> **Було: **"+"_%s_"+"\n"+"**Стало: **"+"_%s_",
				UserMessage,
				m.Content,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/g4OsjhU.png",
			},
			Color:     0xeda15f, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsMessages, embed)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		file.Close()

	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) { // Модуль логування видаленого повідомлення
		cfg, err := ini.Load("servers/" + m.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsMessages := section.Key("CHANNEL_LOGS_MESSAGE_ID").String()
		ChannelLogsVoice := section.Key("CHANNEL_LOGS_VOICE_ID").String()
		ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
		switch {
		case len(ChannelLogsMessages) != 19:
			return
		case len(ChannelLogsVoice) != 19:
			return
		case len(ChannelLogsServer) != 19:
			return
		}
		if m.ChannelID == ChannelLogsMessages || m.ChannelID == ChannelLogsServer || m.ChannelID == ChannelLogsVoice {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		deletedID := m.Message.ID
		UserID := ""
		UserMessage := ""
		ChannelID := ""
		file, err := os.Open("servers/" + m.GuildID + "/message.log")
		if err != nil {
			fmt.Println("Помилка відкриття файлу:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, deletedID) {
				re := regexp.MustCompile(`ID: ([^\s]+)`)
				match := re.FindStringSubmatch(line)
				if len(match) > 1 {
					UserID = match[1]
				}
				re = regexp.MustCompile(`ChannelID: ([^\s]+)`)
				match = re.FindStringSubmatch(line)
				if len(match) > 1 {
					ChannelID = match[1]
				}
				re = regexp.MustCompile(`Text message: ([^|]+)`)
				match = re.FindStringSubmatch(line)
				if len(match) > 1 {
					UserMessage = match[1]
				}
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Помилка при читанні файлу:", err)
			return
		}
		embed := &discordgo.MessageEmbed{
			Title: "Повідомлення видалено!",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "**Канал**",
					Value:  "<#" + ChannelID + ">",
					Inline: true,
				},
				{
					Name:   "**Автор**",
					Value:  "<@" + UserID + ">",
					Inline: true,
				},
			},
			Description: fmt.Sprintf(
				">>> **Текст повідомлення: **\n"+"*%s*",
				UserMessage,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/70d2SGt.png",
			},
			Color:     0xed5f5f, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsMessages, embed)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		file.Close()
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) { // Модуль логування входу/переходу/виходу в голосових каналах
		if userChannels[vs.UserID] == vs.ChannelID {
			return
		}
		cfg, err := ini.Load("servers/" + vs.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsVoice := section.Key("CHANNEL_LOGS_VOICE_ID").String()
		if len(ChannelLogsVoice) != 19 {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

		if vs.VoiceState.ChannelID != "" {
			if userChannels[vs.UserID] == vs.ChannelID {
				return
			}
			if len(userChannels[vs.UserID]) > 10 {
				if vs.ChannelID != userChannels[vs.UserID] {
					embed_run := &discordgo.MessageEmbed{
						Title: "Користувач перейшов в інший голосовий канал",
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "**Старий канал**",
								Value:  "<#" + userChannels[vs.UserID] + ">",
								Inline: true,
							},
							{
								Name:   "**Новий канал**",
								Value:  "<#" + vs.ChannelID + ">",
								Inline: true,
							},
						},
						Description: fmt.Sprintf(
							">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`",
							vs.UserID,
							vs.UserID,
						),
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: "https://i.imgur.com/ARqm68x.png",
						},
						Color:     0xc9c9c9, // Колір (у форматі HEX)
						Timestamp: stringTime,
						Footer: &discordgo.MessageEmbedFooter{
							Text:    vs.Member.User.Username,
							IconURL: vs.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
						},
					}
					_, err = s.ChannelMessageSendEmbed(ChannelLogsVoice, embed_run)
					if err != nil {
						fmt.Println("error getting member:", err)
						return
					}
					userChannels[vs.UserID] = vs.ChannelID
					return
				}
			}
			embed_join := &discordgo.MessageEmbed{
				Title: "Користувач зайшов в голосовий канал",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "**Канал**",
						Value:  "<#" + vs.ChannelID + ">",
						Inline: true,
					},
					{
						Name:   "**Користувач**",
						Value:  "<@" + vs.UserID + ">",
						Inline: true,
					},
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://i.imgur.com/HfR2ekf.png",
				},
				Color:     0x5fed80, // Колір (у форматі HEX)
				Timestamp: stringTime,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    vs.Member.User.Username,
					IconURL: vs.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
				},
			}
			_, err = s.ChannelMessageSendEmbed(ChannelLogsVoice, embed_join)
			if err != nil {
				fmt.Println("error getting member:", err)
				return
			}
			JoinTime := time.Now()
			userTimeJoinVoice[vs.UserID] = strconv.FormatInt(JoinTime.Unix(), 10)
			userChannels[vs.UserID] = vs.ChannelID
		} else {
			channelID := userChannels[vs.UserID]
			embed_leave := &discordgo.MessageEmbed{
				Title: "Користувач вийшов з голосового каналу",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "**Канал**",
						Value:  "<#" + channelID + ">",
						Inline: true,
					},
					{
						Name:   "**Користувач**",
						Value:  "<@" + vs.UserID + ">",
						Inline: true,
					},
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://i.imgur.com/K6wF5SK.png",
				},
				Color:     0xed5f5f, // Колір (у форматі HEX)
				Timestamp: stringTime,
				Footer: &discordgo.MessageEmbedFooter{
					Text:    vs.Member.User.Username,
					IconURL: vs.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
				},
			}
			_, err = s.ChannelMessageSendEmbed(ChannelLogsVoice, embed_leave)
			if err != nil {
				fmt.Println("error getting member:", err)
				return
			}
			LeaveTime := time.Now()
			JoinTime := userTimeJoinVoice[vs.UserID]
			// Переведення JoinTime в int64 (Unix-час)
			JoinTimeUnix, err := strconv.ParseInt(JoinTime, 10, 64)
			if err != nil {
				// Обробка помилки, якщо парсинг невдалося
				fmt.Println("Помилка конвертації JoinTime в int64:", err)
				return
			}
			// Визначення різниці в часі між LeaveTime та JoinTime
			currentTime := (LeaveTime.Unix() - JoinTimeUnix) / 60
			var EXP uint32
			section = cfg.Section("LVL_EXP_USERS")
			valueStr := section.Key(vs.UserID).String()
			parsedEXP, err := strconv.ParseUint(valueStr, 10, 32)
			if err != nil {
				fmt.Println("Помилка конвертації значення рядка в uint32:", err)
				// Обробка помилки, якщо потрібно
				return
			}
			EXP = uint32(parsedEXP)
			EXP += uint32(currentTime)
			EXPStr := strconv.Itoa(int(EXP))
			section.Key(vs.UserID).SetValue(EXPStr)
			basePath := "./servers"
			folderName := vs.GuildID
			directoryPath := filepath.Join(basePath, folderName)
			filePath := filepath.Join(directoryPath, "config.ini")
			err = cfg.SaveTo(filePath)
			if err != nil {
				fmt.Println("Помилка при збереженні у файл:", err)
				return
			}
			delete(userTimeJoinVoice, vs.UserID)
			delete(userChannels, vs.UserID)
		}
	})
	sess.AddHandler(func(s *discordgo.Session, gma *discordgo.GuildMemberAdd) { // Модуль логування надходження користувачів на сервер
		cfg, err := ini.Load("servers/" + gma.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
		if len(ChannelLogsServer) != 19 {
			return
		}

		currentTime := time.Now()
		userTimeJoin[gma.User.ID] = strconv.FormatInt(currentTime.Unix(), 10)

		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		creationTime, err := discordgo.SnowflakeTimestamp(gma.User.ID)
		if err != nil {
			fmt.Println("Помилка отримання дати створення облікового запису:", err)
			return
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач приєднався",
			Description: fmt.Sprintf(
				">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"\n**Створений: **"+"<t:"+"%d"+":R>",
				gma.User.ID,
				gma.User.ID,
				int(creationTime.Unix()),
			),
			Color:     0x1b7ab5, // Колір (у форматі HEX)
			Timestamp: stringTime,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/jxNB6yn.png",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    gma.Member.User.Username,
				IconURL: gma.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
			},
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) { // Модуль логування виходу користувачів з серверу
		cfg, err := ini.Load("servers/" + gmr.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
		if len(ChannelLogsServer) != 19 {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		userTime, err := strconv.ParseInt(userTimeJoin[gmr.User.ID], 10, 64)
		if err != nil {
			fmt.Println("Помилка конвертації строки в int64:", err)
			return
		}
		stringTemp := "<t:" + strconv.FormatInt(userTime, 10) + ":R>"

		if err != nil {
			fmt.Println("Помилка отримання дати створення облікового запису:", err)
			return
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач покинув сервер",
			Description: fmt.Sprintf(
				">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"\n"+"**Приєднався: **"+"%s",
				gmr.User.ID,
				gmr.User.ID,
				stringTemp,
			),
			Color:     0xe3ad62, // Колір (у форматі HEX)
			Timestamp: stringTime,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/iwsJcJn.png",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    gmr.Member.User.Username,
				IconURL: gmr.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
			},
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		delete(userTimeJoin, gmr.User.ID)
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildBanAdd) { // Модуль логування бану користувачів на сервер
		cfg, err := ini.Load("servers/" + gmr.GuildID + "/config.ini")
		if err != nil {
			fmt.Println("Помилка при завантаженні конфігураційного файлу:", err)
			return
		}
		section := cfg.Section("LOGS")
		ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
		if len(ChannelLogsServer) != 19 {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

		if err != nil {
			fmt.Println("Помилка отримання дати створення облікового запису:", err)
			return
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач був забанений",
			Description: fmt.Sprintf(
				">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"\n",
				gmr.User.ID,
				gmr.User.ID,
			),
			Color:     0xeb5079, // Колір (у форматі HEX)
			Timestamp: stringTime,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/MtFRxOr.png",
			},
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	})
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers // Доп. дозволи

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("The bot is online!")

	sc := make(chan os.Signal, 1) // Вимкнення бота CTRL+C
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
