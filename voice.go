package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func VoiceLog(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(vs.GuildID), vs.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Voice == "0" {
		return
	}

	if vs.BeforeUpdate == nil {
		embed := &discordgo.MessageEmbed{
			Title: "The user entered the room",
			Description: fmt.Sprintf(
				">>> **Channel: **"+"<#%s>"+"\n"+"**User: **"+"<@%s>",
				vs.ChannelID, vs.Member.User.ID,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: vs.Member.AvatarURL("256"),
			},
			Color:     0x5fc437, // Колір (у форматі HEX)
			Timestamp: time.Now().Format(time.RFC3339),
		}

		_, err := s.ChannelMessageSendEmbed(rows.Channel_ID_Voice, embed)
		if err != nil {
			Error("error join the room", err)
			return
		}
		return
	}
	if vs.BeforeUpdate.ChannelID == vs.ChannelID {
		return
	}
	if vs.ChannelID == "" {
		embed := &discordgo.MessageEmbed{
			Title: "The user left the room",

			Description: fmt.Sprintf(
				">>> **Channel: **"+"<#%s>"+"\n"+"**User: **"+"<@%s>",
				vs.BeforeUpdate.ChannelID, vs.Member.User.ID,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: vs.Member.AvatarURL("256"),
			},
			Color:     0xc43737, // Колір (у форматі HEX)
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err := s.ChannelMessageSendEmbed(rows.Channel_ID_Voice, embed)
		if err != nil {
			Error("error leave the room", err)
			return
		}
		return

	}
	if vs.BeforeUpdate.ChannelID != vs.ChannelID {
		embed := &discordgo.MessageEmbed{
			Title: "The user moved to the second room",
			Description: fmt.Sprintf(
				">>> **Old room: **"+"<#%s>"+"\n"+"**New room: **"+"<#%s>"+"\n"+"**User: **"+"<@%s>",
				vs.BeforeUpdate.ChannelID, vs.ChannelID, vs.Member.User.ID,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: vs.Member.AvatarURL("256"),
			},
			Color:     0x37c4b8, // Колір (у форматі HEX)
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err := s.ChannelMessageSendEmbed(rows.Channel_ID_Voice, embed)
		if err != nil {
			Error("error move member room", err)
			return
		}
		return
	}
}
