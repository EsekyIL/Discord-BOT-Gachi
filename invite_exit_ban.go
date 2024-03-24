package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

func InvitePeopleToServer(s *discordgo.Session, gma *discordgo.GuildMemberAdd) {
	cfg, err := ini.Load("servers/" + gma.GuildID + "/config.ini")
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при завантаженні конфігураційного файлу: %v", err)
		writer := color.New(color.FgBlue, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
		return
	}
	section := cfg.Section("LOGS")
	ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
	if len(ChannelLogsServer) != 19 {
		return
	}

	currentTime := time.Now()

	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
	creationTime, err := discordgo.SnowflakeTimestamp(gma.User.ID)
	if err != nil {
		fmt.Println("Помилка отримання дати створення облікового запису:", err)
		return
	}
	embed_join := &discordgo.MessageEmbed{
		Title: "Користувач приєднався",
		Description: fmt.Sprintf(
			">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"\n**Створений: **"+"<t:"+"%d"+":R>",
			gma.User.ID,
			gma.User.ID,
			int(creationTime.Unix()),
		),
		Color:     0x1b7ab5, // Колір (у форматі HEX)
		Timestamp: stringTime,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/jxNB6yn.png",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    gma.Member.User.Username,
			IconURL: gma.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
	}
	_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_join)
	if err != nil {
		fmt.Println("error getting member:", err)
		return
	}
}
func ExitPeopleFromServer(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) {
	cfg, err := ini.Load("servers/" + gmr.GuildID + "/config.ini")
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при завантаженні конфігураційного файлу: %v", err)
		writer := color.New(color.FgBlue, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
		return
	}
	section := cfg.Section("LOGS")
	ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
	if len(ChannelLogsServer) != 19 {
		return
	}

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	embed_join := &discordgo.MessageEmbed{
		Title: "Користувач покинув сервер",
		Description: fmt.Sprintf(
			">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"\n",
			gmr.User.ID,
			gmr.User.ID,
		),
		Color:     0xe3ad62, // Колір (у форматі HEX)
		Timestamp: stringTime,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/iwsJcJn.png",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    gmr.Member.User.Username,
			IconURL: gmr.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
	}
	embed_exile := &discordgo.MessageEmbed{
		Title: "Користувач покинув сервер",
		Description: fmt.Sprintf(
			">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"\n"+"***\nПокинув сервер тому що погано себе поводив!***",
			gmr.Member.User.ID,
			gmr.Member.User.ID,
		),
		Color:     0xe3ad62, // Колір (у форматі HEX)
		Timestamp: stringTime,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/iwsJcJn.png",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    gmr.Member.User.Username,
			IconURL: gmr.Member.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
		},
	}
	if gmr.Member.User == nil {
		_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_exile)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	} else {
		_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_join)
		if err != nil {
			fmt.Println("error getting member:", err)
			return
		}
	}
}
func BanUserToServer(s *discordgo.Session, b *discordgo.GuildBanAdd, ba *discordgo.GuildBan) {
	cfg, err := ini.Load("servers/" + b.GuildID + "/config.ini")
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при завантаженні конфігураційного файлу: %v", err)
		writer := color.New(color.FgBlue, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
		return
	}
	section := cfg.Section("LOGS")
	ChannelLogsServer := section.Key("CHANNEL_LOGS_SERVER_ID").String()
	if len(ChannelLogsServer) != 19 {
		return
	}
	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")
	embed_join := &discordgo.MessageEmbed{
		Title: "Користувач був забанений",
		Description: fmt.Sprintf(
			">>> **Користувач: **<@%s>\n**Айді: **"+"`%s`"+"**\nПричина: **"+"`%s`",
			b.User.ID,
			b.User.ID,
			ba.Reason,
		),
		Color:     0xeb5079, // Колір (у форматі HEX)
		Timestamp: stringTime,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/MtFRxOr.png",
		},
	}
	_, err = s.ChannelMessageSendEmbed(ChannelLogsServer, embed_join)
	if err != nil {
		fmt.Println("error getting member:", err)
		return
	}
}
