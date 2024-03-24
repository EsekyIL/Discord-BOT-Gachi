package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

func VoiceLog(s *discordgo.Session, vs *discordgo.VoiceStateUpdate, userChannels *map[string]string, userTimeJoinVoice *map[string]string) {
	if channelID, ok := (*userChannels)[vs.UserID]; ok && channelID == vs.ChannelID {
		return
	}
	cfg, err := ini.Load("servers/" + vs.GuildID + "/config.ini")
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при завантаженні конфігураційного файлу: %v", err)
		writer := color.New(color.FgBlue, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
		return
	}
	section := cfg.Section("LOGS")
	ChannelLogsVoice := section.Key("CHANNEL_LOGS_VOICE_ID").String()
	if len(ChannelLogsVoice) != 19 {
		return
	}
	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
	if vs.VoiceState.ChannelID != "" {
		if (*userChannels)[vs.UserID] == vs.ChannelID {
			return
		}
		if len((*userChannels)[vs.UserID]) > 10 {
			if vs.ChannelID != (*userChannels)[vs.UserID] {
				embed_run := &discordgo.MessageEmbed{
					Title: "Користувач перейшов в інший голосовий канал",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "**Старий канал**",
							Value:  "<#" + (*userChannels)[vs.UserID] + ">",
							Inline: true,
						},
						{
							Name:   "**Новий канал**",
							Value:  "<#" + vs.ChannelID + ">",
							Inline: true,
						},
					},
					Description: fmt.Sprintf(
						">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`",
						vs.UserID,
						vs.UserID,
					),
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://i.imgur.com/ARqm68x.png",
					},
					Color:     0xc9c9c9, // Колір (у форматі HEX)
					Timestamp: stringTime,
					Footer: &discordgo.MessageEmbedFooter{
						Text:    vs.Member.User.Username,
						IconURL: vs.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
					},
				}
				_, err = s.ChannelMessageSendEmbed(ChannelLogsVoice, embed_run)
				if err != nil {
					fmt.Println("error getting member:", err)
					return
				}
				(*userChannels)[vs.UserID] = vs.ChannelID
				return
			}
		}
		embed_join := &discordgo.MessageEmbed{
			Title: "Користувач зайшов в голосовий канал",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "**Канал**",
					Value:  "<#" + vs.ChannelID + ">",
					Inline: true,
				},
				{
					Name:   "**Користувач**",
					Value:  "<@" + vs.UserID + ">",
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/HfR2ekf.png",
			},
			Color:     0x5fed80, // Колір (у форматі HEX)
			Timestamp: stringTime,
			Footer: &discordgo.MessageEmbedFooter{
				Text:    vs.Member.User.Username,
				IconURL: vs.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
			},
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsVoice, embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		JoinTime := time.Now()
		(*userTimeJoinVoice)[vs.UserID] = strconv.FormatInt(JoinTime.Unix(), 10)
		(*userChannels)[vs.UserID] = vs.ChannelID
	} else {
		channelID := (*userChannels)[vs.UserID]
		embed_leave := &discordgo.MessageEmbed{
			Title: "Користувач вийшов з голосового каналу",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "**Канал**",
					Value:  "<#" + channelID + ">",
					Inline: true,
				},
				{
					Name:   "**Користувач**",
					Value:  "<@" + vs.UserID + ">",
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.imgur.com/K6wF5SK.png",
			},
			Color:     0xed5f5f, // Колір (у форматі HEX)
			Timestamp: stringTime,
			Footer: &discordgo.MessageEmbedFooter{
				Text:    vs.Member.User.Username,
				IconURL: vs.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
			},
		}
		_, err = s.ChannelMessageSendEmbed(ChannelLogsVoice, embed_leave)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
		LeaveTime := time.Now()
		JoinTime := (*userTimeJoinVoice)[vs.UserID]
		// Переведення JoinTime в int64 (Unix-час)
		JoinTimeUnix, err := strconv.ParseInt(JoinTime, 10, 64)
		if err != nil {
			// Обробка помилки, якщо парсинг невдалося
			fmt.Println("Помилка конвертації JoinTime в int64:", err)
			return
		}
		// Визначення різниці в часі між LeaveTime та JoinTime
		currentTime := (LeaveTime.Unix() - JoinTimeUnix) / 60
		var EXP uint32
		section = cfg.Section("LVL_EXP_USERS")
		valueStr := section.Key(vs.UserID).String()
		parsedEXP, err := strconv.ParseUint(valueStr, 10, 32)
		if err != nil {
			fmt.Println("Помилка конвертації значення рядка в uint32:", err)
			// Обробка помилки, якщо потрібно
			return
		}
		EXP = uint32(parsedEXP)
		EXP += uint32(currentTime)
		EXPStr := strconv.Itoa(int(EXP))
		section.Key(vs.UserID).SetValue(EXPStr)
		basePath := "./servers"
		folderName := vs.GuildID
		directoryPath := filepath.Join(basePath, folderName)
		filePath := filepath.Join(directoryPath, "config.ini")
		err = cfg.SaveTo(filePath)
		if err != nil {
			fmt.Println("Помилка при збереженні у файл:", err)
			return
		}
		delete((*userTimeJoinVoice), vs.UserID)
		delete((*userChannels), vs.UserID)
	}
}
