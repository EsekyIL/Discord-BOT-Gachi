package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func MsgUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	channel_log_msgID, _ := SelectDB("channel_log_msgID", m.GuildID)
	if channel_log_msgID == 0 {
		return
	}

	currentTime := time.Now().Format(time.RFC3339)

	embed := &discordgo.MessageEmbed{
		Title: "Message updated",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "*Was**",
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
		Color:     0x37c4b8, // Колір (у форматі HEX)
		Timestamp: currentTime,
	}

	_, err := s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_msgID), embed)
	if err != nil {
		Error("error message update", err)
		return
	}

}
func MsgDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	channel_log_msgID, _ := SelectDB("channel_log_msgID", m.GuildID)
	if channel_log_msgID == 0 || m.BeforeDelete == nil {
		return
	}

	currentTime := time.Now().Format(time.RFC3339)

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
		Timestamp: currentTime,
	}
	_, err := s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_msgID), embed)
	if err != nil {
		Error("error message deleted", err)
		return
	}
}
