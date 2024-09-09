package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func VoiceLog(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	channel_log_voiceID, lang := SelectDB("channel_log_voiceID", vs.GuildID)
	if channel_log_voiceID == 0 {
		return
	}

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	trs := getTranslation(lang)

	if vs.BeforeUpdate == nil {
		embed := &discordgo.MessageEmbed{
			Title: trs.UserJoinVoice,
			Description: fmt.Sprintf(
				">>> **%s: **"+"<#%s>"+"\n"+"**%s: **"+"<@%s>",
				trs.Channel, vs.ChannelID, trs.User, vs.Member.User.ID,
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
			Title: trs.UserLeftVoice,

			Description: fmt.Sprintf(
				">>> **%s: **"+"<#%s>"+"\n"+"**%s: **"+"<@%s>",
				trs.Channel, vs.BeforeUpdate.ChannelID, trs.User, vs.Member.User.ID,
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
			Title: trs.UserMovedVoice,
			Description: fmt.Sprintf(
				">>> **%s: **"+"<#%s>"+"\n"+"**%s: **"+"<#%s>"+"\n"+"**%s: **"+"<@%s>",
				trs.OldRoom, vs.BeforeUpdate.ChannelID, trs.NewRoom, vs.ChannelID, trs.User, vs.Member.User.ID,
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
