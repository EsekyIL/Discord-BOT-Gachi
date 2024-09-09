package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func MsgUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	channel_log_msgID, lang := SelectDB("channel_log_msgID", m.GuildID)
	if channel_log_msgID == 0 {
		return
	}

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	trs := getTranslation(lang)

	embed := &discordgo.MessageEmbed{
		Title: trs.MessageUpdated,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("**%s**", trs.Was),
				Value:  m.BeforeUpdate.Content,
				Inline: true,
			},
			{
				Name:   fmt.Sprintf("**%s**", trs.NowIs),
				Value:  m.Content,
				Inline: true,
			},
		},
		Description: fmt.Sprintf(
			">>> **%s: **"+"<#%s>"+"\n"+"**%s: **"+"<@%s>"+"\n"+"**%s: **"+"[%s](https://discord.com/channels/%s/%s/%s)",
			trs.Channel, m.ChannelID, trs.Author, m.Author.ID, trs.MessageID, m.Message.ID, m.GuildID, m.ChannelID, m.ID,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    m.Author.Username,
			IconURL: m.Author.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
		Color:     0x37c4b8, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}

	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_msgID), embed)

}
func MsgDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	channel_log_msgID, lang := SelectDB("channel_log_msgID", m.GuildID)
	if channel_log_msgID == 0 || m.BeforeDelete == nil {
		return
	}

	currentTime := time.Now().Format("2006-01-02T15:04:05.999Z07:00")

	trs := getTranslation(lang)

	embed := &discordgo.MessageEmbed{
		Title: trs.MessageDeleted,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("**%s**", trs.MessageContent),
				Value:  "*" + m.BeforeDelete.Content + "*",
				Inline: false,
			},
		},
		Description: fmt.Sprintf(
			">>> **%s: **"+"<#%s>"+"\n"+"**%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`",
			trs.Channel, m.ChannelID, trs.Author, m.BeforeDelete.Author.ID, trs.MessageID, m.Message.ID,
		),
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: currentTime,
	}
	_, err := s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_msgID), embed)
	if err != nil {
		fmt.Println("error getting member:", err)
		return
	}
}
func AllMessageDeletedInChannel() {

}
