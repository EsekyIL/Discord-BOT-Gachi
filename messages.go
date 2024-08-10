package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

type LogStruct struct {
	MsgID     int    `json:"ID"`
	Name      string `json:"NAME"`
	Msg       string `json:"MSG"`
	ID        int    `json:"MSG_ID"`
	ChannelID int    `json:"CHANNEL_ID"`
	Status    string `json:"STATUS"`
}

func MessageSaveToLog(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	if m.ChannelID == ChannelLogsMessages || m.ChannelID == ChannelLogsVoice || m.ChannelID == ChannelLogsServer {
		return
	} else {
		logFilePath := filepath.Join("servers", m.GuildID, "message.json")

		// Зчитування існуючих даних
		var logs []LogStruct

		if _, err := os.Stat(logFilePath); err == nil {
			fileData, err := os.ReadFile(logFilePath)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}

			// Якщо файл не порожній, пробуємо розпарсити JSON
			if len(fileData) > 0 {
				if err := json.Unmarshal(fileData, &logs); err != nil {
					fmt.Println("Error unmarshaling JSON:", err)
					return
				}
			}
		}

		// Додаємо новий запис
		MSG_ID, _ := strconv.Atoi(m.Message.ID)
		AUTHOR_ID, _ := strconv.Atoi(m.Author.ID)
		CHANNEL_ID, _ := strconv.Atoi(m.ChannelID)

		logs = append(logs, LogStruct{
			MsgID:     MSG_ID,
			Name:      m.Author.Username,
			Msg:       m.Content,
			ID:        AUTHOR_ID,
			ChannelID: CHANNEL_ID,
			Status:    "Нове",
		})

		// Серіалізація масиву в JSON
		data, err := json.MarshalIndent(logs, "", "  ") // Використовуємо MarshalIndent для читабельності
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		// Запис оновленого масиву в файл
		if err := os.WriteFile(logFilePath, data, 0666); err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
}
func MessageUpdateToLog(s *discordgo.Session, m *discordgo.MessageUpdate) {
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
	if m.ChannelID == ChannelLogsMessages || m.ChannelID == ChannelLogsServer || m.ChannelID == ChannelLogsVoice {
		return
	}

	switch {
	case len(ChannelLogsMessages) != 19:
		return
	case len(ChannelLogsVoice) != 19:
		return
	case len(ChannelLogsServer) != 19:
		return
	}

	logFilePath := filepath.Join("servers", m.GuildID, "message.json")

	var logs []LogStruct

	if _, err := os.Stat(logFilePath); err == nil {
		fileData, err := os.ReadFile(logFilePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		// Якщо файл не порожній, пробуємо розпарсити JSON
		if len(fileData) > 0 {
			if err := json.Unmarshal(fileData, &logs); err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				return
			}
		}
	}

	MSG_ID, _ := strconv.Atoi(m.Message.ID)
	AUTHOR_ID, _ := strconv.Atoi(m.Author.ID)
	CHANNEL_ID, _ := strconv.Atoi(m.ChannelID)
	var UserInfo LogStruct

	for i, log := range logs {
		if log.MsgID == MSG_ID {
			UserInfo = log
			logs[i] = LogStruct{
				MsgID:     MSG_ID,
				Name:      m.Author.Username,
				Msg:       m.Content,
				ID:        AUTHOR_ID,
				ChannelID: CHANNEL_ID,
				Status:    "Оновлене",
			}
			break
		}
	}
	data, err := json.MarshalIndent(logs, "", "  ") // Використовуємо MarshalIndent для читабельності
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Запис оновленого масиву в файл
	if err := os.WriteFile(logFilePath, data, 0666); err != nil {
		fmt.Println("Error writing to file:", err)
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
			UserInfo.Msg,
			m.Content,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    m.Author.Username,
			IconURL: m.Author.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
		Color:     0xeda15f, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, err = s.ChannelMessageSendEmbed(ChannelLogsMessages, embed)
	if err != nil {
		fmt.Println("error getting member:", err)
		return
	}
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
