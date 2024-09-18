package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func MsgUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.BeforeUpdate == nil {
		return
	}
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(m.GuildID), m.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Message == "0" {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Message updated",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "**Was**",
				Value:  m.BeforeUpdate.Content,
				Inline: true,
			},
			{
				Name:   "**Now is**",
				Value:  m.Content,
				Inline: true,
			},
		},
		Description: fmt.Sprintf(
			">>> **Channel: **"+"<#%s>"+"\n"+"**Author: **"+"<@%s>"+"\n"+"**Message ID: **"+"[%s](https://discord.com/channels/%s/%s/%s)",
			m.ChannelID, m.Author.ID, m.Message.ID, m.GuildID, m.ChannelID, m.ID,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    m.Author.Username,
			IconURL: m.Author.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
		Color:     0xc4b137, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Message, embed)
	if err != nil {
		Error("error message update", err)
		return
	}

}
func MsgDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(m.GuildID), m.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Message == "0" || m.BeforeDelete == nil {
		return
	}
	embed := &discordgo.MessageEmbed{
		Title: "Message deleted",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "**Content**",
				Value:  "*" + m.BeforeDelete.Content + "*",
				Inline: false,
			},
		},
		Description: fmt.Sprintf(
			">>> **Channel: **"+"<#%s>"+"\n"+"**Author: **"+"<@%s>"+"\n"+"**Message ID: **"+"`%s`",
			m.ChannelID, m.BeforeDelete.Author.ID, m.Message.ID,
		),
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Message, embed)
	if err != nil {
		Error("error message deleted", err)
		return
	}
}
