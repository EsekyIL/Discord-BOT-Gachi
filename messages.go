package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
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
				Name:   "**Було**",
				Value:  m.BeforeUpdate.Content,
				Inline: true,
			},
			{
				Name:   "**Стало**",
				Value:  m.Content,
				Inline: true,
			},
		},
		Description: fmt.Sprintf(
			">>> **Канал: **"+"<#%s>"+"\n"+"**Автор: **"+"<@%s>"+"\n"+"**Айді повідомлення: **"+"`%s`",
			m.ChannelID, m.Author.ID, m.Message.ID,
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
func MsgDelete(s *discordgo.Session, m *discordgo.MessageDelete, database *sql.DB) {
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
		Title: "Видалене повідомлення",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "**Вміст повідомлення**",
				Value:  "*" + m.BeforeDelete.Content + "*",
				Inline: false,
			},
		},
		Description: fmt.Sprintf(
			">>> **Канал: **"+"<#%s>"+"\n"+"**Автор: **"+"<@%s>"+"\n"+"**Айді повідомлення: **"+"`%s`",
			m.ChannelID, m.BeforeDelete.Author.ID, m.Message.ID,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/70d2SGt.png",
		},
		Color:     0xed5f5f, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, err = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_msgID), embed)
	if err != nil {
		fmt.Println("error getting member:", err)
		return
	}
}
func AllMessageDeletedInChannel() {

}
