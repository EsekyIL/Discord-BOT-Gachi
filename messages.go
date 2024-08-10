package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

func MsgUpdate(s *discordgo.Session, m *discordgo.MessageUpdate, database *sql.DB) {
	rows, err := database.Query("SELECT id, channel_log_msgID FROM servers WHERE id = ?", m.GuildID)
	if err != nil {
		Error("Щось сталось", err)
		return // Зупиняємо виконання у разі помилки
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var id int
	var channel_log_msgID int

	if rows.Next() {
		err := rows.Scan(&id, &channel_log_msgID)
		if err != nil {
			Error("Failed to scan the row", err)
			return
		}
	} else {
		if err := rows.Err(); err != nil {
			Error("Failed during iteration over rows", err)
		}
		return
	}

	if id == 0 {
		return
	}

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

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
			m.BeforeUpdate.Content,
			m.Content,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    m.Author.Username,
			IconURL: m.Author.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
		Color:     0xeda15f, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}

	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_msgID), embed)

}
func MessageDeleteLog(s *discordgo.Session, m *discordgo.MessageDelete) {
	cfg, err := ini.Load("servers/" + m.GuildID + "/config.ini")
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при завантаженні конфігураційного файлу: %v", err)
		writer := color.New(color.FgBlue, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
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
}
func AllMessageDeletedInChannel() {

}
