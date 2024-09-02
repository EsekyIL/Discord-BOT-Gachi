package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func VoiceLog(s *discordgo.Session, vs *discordgo.VoiceStateUpdate, database *sql.DB) {
	channel_log_voiceID := SelectDB("channel_log_voiceID", vs.GuildID, database)
	if channel_log_voiceID == 0 {
		return
	}

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	if vs.BeforeUpdate == nil {
		embed := &discordgo.MessageEmbed{
			Title: "Користувач зайшов у кімнату",
			Description: fmt.Sprintf(
				">>> **Канал: **"+"<#%s>"+"\n"+"**Користувач: **"+"<@%s>",
				vs.ChannelID, vs.Member.User.ID,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: vs.Member.AvatarURL("256"),
			},
			Color:     0x5fc437, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}

		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_voiceID), embed)
		return
	}
	if vs.BeforeUpdate.ChannelID == vs.ChannelID {
		return
	}
	if vs.ChannelID == "" {
		embed := &discordgo.MessageEmbed{
			Title: "Користувач вийшов з кімнати",

			Description: fmt.Sprintf(
				">>> **Канал: **"+"<#%s>"+"\n"+"**Користувач: **"+"<@%s>",
				vs.BeforeUpdate.ChannelID, vs.Member.User.ID,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: vs.Member.AvatarURL("256"),
			},
			Color:     0xc43737, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_voiceID), embed)
		return

	}
	if vs.BeforeUpdate.ChannelID != vs.ChannelID {
		embed := &discordgo.MessageEmbed{
			Title: "Користувач перейшов у другу кімнату",
			Description: fmt.Sprintf(
				">>> **Стара кімната: **"+"<#%s>"+"\n"+"**Нова кімната: **"+"<#%s>"+"\n"+"**Користувач: **"+"<@%s>",
				vs.BeforeUpdate.ChannelID, vs.ChannelID, vs.Member.User.ID,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: vs.Member.AvatarURL("256"),
			},
			Color:     0x37c4b8, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_voiceID), embed)
		return
	}
}
