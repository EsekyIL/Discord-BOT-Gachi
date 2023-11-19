package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *log.Logger

func main() {
	const Token = "MTE2MDE3NTg5NTQ3NTEzODYxMQ.GblEFM.v-JGilyUhGd9g_ixkBAg3JNzV2ryFPy60afouQ"
	const commandPrefix = "!"
	addroleID := "1161309104283865100"
	addrolelvlID := "1161310698975002654"
	var userChannels map[string]string
	userChannels = make(map[string]string)

	l := &lumberjack.Logger{
		Filename:   "path/logs/message.log",
		MaxSize:    500, // мегабайти
		MaxBackups: 3,
		MaxAge:     1, // дні
	}
	logger = log.New(l, "", log.LstdFlags)
	sess, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal(err)
	}
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		if m.MessageID == "1161369411710615623" {
			// Перевіряємо, чи це потрібна реакція (emoji)
			if m.Emoji.Name == "🎮" {
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
					if role == addrolelvlID {
						hasRole = true
						break
					}
				}
				if hasRole {
					// Користувач має певну роль, надсилаємо йому приватне повідомлення
					currentTime := time.Now()
					stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
					channel, err := s.UserChannelCreate(userID)
					if err != nil {
						fmt.Println("error creating channel:", err)
						return
					}
					// Надсилання приватного повідомлення
					embed := &discordgo.MessageEmbed{
						Title:       "⚠️ Помилка! ⚠️",
						Description: "Вам вже видана роль! Якщо ролі немає - зверніться до адміністрації серверу",
						Color:       0xf5b507, // Колір (у форматі HEX)
						Timestamp:   stringTime,
					}
					_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
					if err != nil {
						fmt.Println("error sending message:", err)
						return
					}
				} else {
					err = s.GuildMemberRoleAdd(m.GuildID, userID, addrolelvlID)
					if err != nil {
						fmt.Println("error adding role,", err)
						return
					}
					err = s.GuildMemberRoleAdd(m.GuildID, userID, addroleID)
					if err != nil {
						fmt.Println("error adding role,", err)
						return
					}
				}
			}
		}

	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.HasPrefix(m.Content, commandPrefix+"c-message") {
			// Перевірка, чи користувач є адміністратором
			member, err := s.GuildMember(m.GuildID, m.Author.ID)
			if err != nil {
				fmt.Println("Помилка отримання інформації про користувача:", err)
				return
			}

			for _, roleID := range member.Roles {
				role, err := s.State.Role(m.GuildID, roleID)
				if err != nil {
					fmt.Println("Помилка отримання інформації про роль:", err)
					continue
				}
				if role.Permissions&discordgo.PermissionAdministrator != 0 {
					// Користувач є адміністратором, викликаємо команду налаштувань

					return
				}
			}

			// Користувач не є адміністратором, можна вивести повідомлення про відмову
			s.ChannelMessageSend(m.ChannelID, "Ви не маєте права викликати цю команду.")
		}
		if m.ChannelID == "1161397001817169980" || m.ChannelID == "1161397893622661240" || m.ChannelID == "1161398323056488589" {
			return
		} else {
			logger.Println("Text message: " + m.Content + " | " + "Nickname: " + m.Author.Username + " | " + "ID: " + m.Author.ID + " | " + "messageID: " + m.Message.ID + " | " + "ChannelID: " + m.ChannelID)
		}
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.ChannelID == "1161397001817169980" || m.ChannelID == "1161397893622661240" || m.ChannelID == "1161398323056488589" {
			return
		}
		if m.Author == nil || m.Author.Bot {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		MessageUpdateID := m.Message.ID
		UserMessage := ""
		file, err := os.OpenFile("path/logs/message.log", os.O_RDWR, 0644)
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
				{
					Name:   "**Було**",
					Value:  UserMessage,
					Inline: false,
				},
				{
					Name:   "**Стало**",
					Value:  m.Content,
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/g4OsjhU.png",
			},
			Color:     0xeda15f, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, err = s.ChannelMessageSendEmbed("1161397001817169980", embed)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		file.Close()

	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		if m.ChannelID == "1161397001817169980" || m.ChannelID == "1161397893622661240" || m.ChannelID == "1161398323056488589" {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		deletedID := m.Message.ID
		UserID := ""
		UserMessage := ""
		ChannelID := ""
		file, err := os.Open("path/logs/message.log")
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
				{
					Name:   "Текст повідомлення",
					Value:  UserMessage,
					Inline: false,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/70d2SGt.png",
			},
			Color:     0xed5f5f, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, err = s.ChannelMessageSendEmbed("1161397001817169980", embed)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		file.Close()
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		if userChannels[vs.UserID] == vs.ChannelID {
			return
		}
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		if vs.ChannelID == "" {
			channelID := userChannels[vs.UserID]
			embed1 := &discordgo.MessageEmbed{
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
			_, err = s.ChannelMessageSendEmbed("1161397893622661240", embed1)
			if err != nil {
				fmt.Println("error getting member:", err)
				return
			}
			delete(userChannels, vs.UserID)
		} else {
			embed2 := &discordgo.MessageEmbed{
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
			if len(userChannels[vs.UserID]) > 10 {
				if vs.ChannelID != userChannels[vs.UserID] {
					embed3 := &discordgo.MessageEmbed{
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
							{
								Name:   "**Користувач**",
								Value:  "<@" + vs.UserID + ">",
								Inline: false,
							},
						},
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
					_, err = s.ChannelMessageSendEmbed("1161397893622661240", embed3)
					if err != nil {
						fmt.Println("error getting member:", err)
						return
					}
					return
				}
			}
			_, err = s.ChannelMessageSendEmbed("1161397893622661240", embed2)
			if err != nil {
				fmt.Println("error getting member:", err)
				return
			}
			userChannels[vs.UserID] = vs.ChannelID
		}
	})
	sess.AddHandler(func(s *discordgo.Session, gma *discordgo.GuildMemberAdd) {
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		creationTime, err := discordgo.SnowflakeTimestamp(gma.User.ID)
		years := time.Since(creationTime).Hours() / 24 / 365
		days := time.Since(creationTime).Hours() / 24
		if err != nil {
			fmt.Println("Помилка отримання дати створення облікового запису:", err)
			return
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач приєднався",
			Description: fmt.Sprintf(
				"**Користувач: **<@%s>\n**Айді: **%s\n**Створений: **%.2f років (%.0f днів)",
				gma.User.ID,
				gma.User.ID,
				years,
				days,
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
		_, err = s.ChannelMessageSendEmbed("1161398323056488589", embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) {
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

		if err != nil {
			fmt.Println("Помилка отримання дати створення облікового запису:", err)
			return
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач покинув сервер",
			Description: fmt.Sprintf(
				"**Користувач: **<@%s>\n**Айді: **%s\n",
				gmr.User.ID,
				gmr.User.ID,
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
		_, err = s.ChannelMessageSendEmbed("1161398323056488589", embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildBanAdd) {
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

		if err != nil {
			fmt.Println("Помилка отримання дати створення облікового запису:", err)
			return
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач був забанений",
			Description: fmt.Sprintf(
				"**Користувач: **<@%s>\n**Айді: **%s\n",
				gmr.User.ID,
				gmr.User.ID,
			),
			Color:     0xeb5079, // Колір (у форматі HEX)
			Timestamp: stringTime,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/MtFRxOr.png",
			},
		}
		_, err = s.ChannelMessageSendEmbed("1161398323056488589", embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	})
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("The bot is online!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
