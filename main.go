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
	Token := "MTE2MDE3NTg5NTQ3NTEzODYxMQ.GblEFM.v-JGilyUhGd9g_ixkBAg3JNzV2ryFPy60afouQ"
	guildID := "965014140357853285"
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
					err = s.GuildMemberRoleAdd(guildID, userID, addrolelvlID)
					if err != nil {
						fmt.Println("error adding role,", err)
						return
					}
					err = s.GuildMemberRoleAdd(guildID, userID, addroleID)
					if err != nil {
						fmt.Println("error adding role,", err)
						return
					}
				}
			}
		}

	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
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
				}
				_, err := file.Seek(-int64(len(line)), io.SeekStart)
				if err != nil {
					fmt.Println("error seeking:", err)
					return
				}
				_, err = file.WriteString("Text message: " + m.Content + " | " + "Nickname: " + m.Author.Username + " | " + "ID: " + m.Author.ID + " | " + "messageID: " + m.Message.ID + " | " + "ChannelID: " + m.ChannelID)
				if err != nil {
					fmt.Println("error writing:", err)
					return
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
				&discordgo.MessageEmbedField{
					Name:   "**Канал**",
					Value:  "<#" + m.ChannelID + ">",
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "**Автор**",
					Value:  "<@" + m.Author.ID + ">",
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "**Було**",
					Value:  UserMessage,
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name:   "**Стало**",
					Value:  m.Content,
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://raw.githubusercontent.com/EsekyIL/Discord-BOT-Gachi/main/reload.png?token=GHSAT0AAAAAACIYYIULUQNDTQQIOJRJOUCCZJKVKTQ",
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
				&discordgo.MessageEmbedField{
					Name:   "**Канал**",
					Value:  "<#" + ChannelID + ">",
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "**Автор**",
					Value:  "<@" + UserID + ">",
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Текст повідомлення",
					Value:  UserMessage,
					Inline: false,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://raw.githubusercontent.com/EsekyIL/Discord-BOT-Gachi/main/delete.png?token=GHSAT0AAAAAACIYYIUKAWFH55GXSFWBGIF2ZJKVA4A",
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
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
		if vs.ChannelID == "" {
			channelID := userChannels[vs.UserID]
			embed1 := &discordgo.MessageEmbed{
				Title: "Користувач вийшов з голосового каналу",
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "**Канал**",
						Value:  "<#" + channelID + ">",
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "**Користувач**",
						Value:  "<@" + vs.UserID + ">",
						Inline: true,
					},
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://raw.githubusercontent.com/EsekyIL/Discord-BOT-Gachi/main/leave.png?token=GHSAT0AAAAAACIYYIUKITA63UEC64H6UCIAZJKVUVQ",
				},
				Color:     0xed5f5f, // Колір (у форматі HEX)
				Timestamp: stringTime,
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
					&discordgo.MessageEmbedField{
						Name:   "**Канал**",
						Value:  "<#" + vs.ChannelID + ">",
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "**Користувач**",
						Value:  "<@" + vs.UserID + ">",
						Inline: true,
					},
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://raw.githubusercontent.com/EsekyIL/Discord-BOT-Gachi/main/join.png?token=GHSAT0AAAAAACIYYIULMIVPGR6SUCNDUQYKZJKV2WA",
				},
				Color:     0x5fed80, // Колір (у форматі HEX)
				Timestamp: stringTime,
			}
			if len(userChannels[vs.UserID]) > 10 {
				if vs.ChannelID != userChannels[vs.UserID] {
					embed3 := &discordgo.MessageEmbed{
						Title: "Користувач перейшов в інший голосовий канал",
						Fields: []*discordgo.MessageEmbedField{
							&discordgo.MessageEmbedField{
								Name:   "**Старий канал**",
								Value:  "<#" + userChannels[vs.UserID] + ">",
								Inline: true,
							},
							&discordgo.MessageEmbedField{
								Name:   "**Новий канал**",
								Value:  "<#" + vs.ChannelID + ">",
								Inline: true,
							},
							&discordgo.MessageEmbedField{
								Name:   "**Користувач**",
								Value:  "<@" + vs.UserID + ">",
								Inline: false,
							},
						},
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: "https://raw.githubusercontent.com/EsekyIL/Discord-BOT-Gachi/main/running.png?token=GHSAT0AAAAAACIYYIUKU4YJQ7AFVM6ACRZOZJKV7HA",
						},
						Color:     0xc9c9c9, // Колір (у форматі HEX)
						Timestamp: stringTime,
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
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

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
