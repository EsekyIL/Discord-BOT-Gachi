package main

import (
	"bufio"
	"fmt"
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

	l := &lumberjack.Logger{
		Filename:   "path/logs/message.log",
		MaxSize:    500, // мегабайти
		MaxBackups: 3,
		MaxAge:     1, // дні
	}

	_, err := l.Write([]byte("test\n"))
	if err != nil {
		log.Fatal(err)
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
		logger.Println(fmt.Sprintf("Message created: %s", "Text message: "+m.Content+" | "+"Nickname: "+m.Author.Username+" | "+"ID: "+m.Author.ID+" | "+"messageID: "+m.Message.ID+" | "+"ChannelID: "+m.ChannelID))
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
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
				fmt.Println(line)
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
		currentTime := time.Now()
		stringTime := currentTime.Format("2006-01-02 15:04:05")
		embed := &discordgo.MessageEmbed{
			Title: "📩 Видалене повідомлення 📩",
			Description: "\n" + UserMessage + "" +
				"\n\n**Канал**" + "\n" + "<#" + ChannelID + ">" +
				"\n" + "**Автор**" + "\n" + "<@" + UserID + ">" +
				"\n\n" + "***" + stringTime + "***",
			Color: 0xed5f5f, // Колір (у форматі HEX)
		}
		_, err = s.ChannelMessageSendEmbed("1161397001817169980", embed)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	})
	/*embed := &discordgo.MessageEmbed{
		Title:       "🎖️ Звання серверу 🎖️",
		Description: "",
		Color:       0x00ff00, // Колір (у форматі HEX)
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		fmt.Println("error getting member:", err)
		return
	}*/
	/*err = s.MessageReactionAdd(m.ChannelID, "1161369411710615623", "🎮")
	  		if err != nil {
	      		fmt.Println("error adding reaction:", err)
	  		}*/

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
