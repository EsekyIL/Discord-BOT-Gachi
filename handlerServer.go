package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func InvCreate(s *discordgo.Session, ic *discordgo.InviteCreate, database *sql.DB) {
	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID := SelectDB("channel_log_serverID", ic.GuildID, database)
	if channel_log_serverID == 0 {
		return
	}

	MaxUses := strconv.Itoa(ic.MaxUses)
	if MaxUses == "0" {
		MaxUses = "♾"
	}

	expiresAt := time.Now().Unix() + int64(ic.MaxAge)
	if ic.MaxAge == 0 {
		expiresAt = 2147483647
	}
	embed := &discordgo.MessageEmbed{
		Title: "Створено запрошення",
		Description: fmt.Sprintf(
			">>> **Код: **"+"`%s`"+"\n"+"**Канал: **"+"<#%s>"+"\n"+"**Термін дії: **"+"<t:%d:R>"+"\n"+"**К-сть користувачів: **"+"%s",
			ic.Code, ic.ChannelID, expiresAt, MaxUses,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ic.Inviter.Username,
			IconURL: ic.Inviter.AvatarURL("256"),
		},
		Color:     0x9d7ff5, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
