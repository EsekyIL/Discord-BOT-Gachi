package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var cache = make(map[string]string)
var cacheMutex sync.RWMutex // Mutex для синхронізації доступу до кешу

func getCreationTime(userID string) time.Time {
	// Перетворюємо ID на ціле число
	id, _ := strconv.ParseInt(userID, 10, 64)

	// Вираховуємо кількість наносекунд з епохи Unix
	timestamp := (id >> 22) + 1420070400000

	// Перетворюємо на тип time.Time
	return time.Unix(0, timestamp*int64(time.Millisecond))
}

func InvCreate(s *discordgo.Session, ic *discordgo.InviteCreate) {
	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID, lang := SelectDB("channel_log_serverID", ic.GuildID)
	if channel_log_serverID == 0 {
		return
	}

	trs := getTranslation(lang)

	MaxUses := strconv.Itoa(ic.MaxUses)
	if MaxUses == "0" {
		MaxUses = "♾"
	}

	expiresAt := time.Now().Unix() + int64(ic.MaxAge)
	if ic.MaxAge == 0 {
		expiresAt = 2147483647
	}
	embed := &discordgo.MessageEmbed{
		Title: trs.InviteCreated,
		Description: fmt.Sprintf(
			">>> **%s: **"+"`%s`"+"\n"+"**%s: **"+"<#%s>"+"\n"+"**%s: **"+"<t:%d:R>"+"\n"+"**%s: **"+"%s",
			trs.Code, ic.Code, trs.Channel, ic.ChannelID, trs.ValidityPeriod, expiresAt, trs.CountUser, MaxUses,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ic.Inviter.Username,
			IconURL: ic.Inviter.AvatarURL("256"),
		},
		Color:     0x37c4b8, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserJoin(s *discordgo.Session, gma *discordgo.GuildMemberAdd) {
	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID, lang := SelectDB("channel_log_serverID", gma.GuildID)
	if channel_log_serverID == 0 {
		return
	}

	trs := getTranslation(lang)
	userCreatedAt := getCreationTime(gma.User.ID)

	embed := &discordgo.MessageEmbed{
		Title: trs.NewUser,
		Description: fmt.Sprintf(
			">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"**%s: **"+"<t:%d:R>"+"\n",
			trs.User, gma.User.ID, trs.ID, gma.User.ID, trs.Created, int64(userCreatedAt.Unix()),
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: gma.AvatarURL("256"),
		},
		Color:     0x5fc437, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserExit(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) {

	cacheKey := fmt.Sprintf("%s:%s", gmr.GuildID, gmr.User.ID)

	cacheMutex.RLock()
	BeforeEntry, exists := cache[cacheKey]
	cacheMutex.RUnlock()

	if !exists {
		BeforeEntry = ""
	}

	AuditLog, err := s.GuildAuditLog(gmr.GuildID, "", "", 20, 1)
	if err != nil {
		Error("AUDIT", err)
		return
	}

	var kick bool
	var UserID string
	var Reason string
	var ActionType string

	for _, entry := range AuditLog.AuditLogEntries {
		UserID = entry.UserID
		Reason = entry.Reason

		if BeforeEntry == entry.ID && BeforeEntry > "" {
			kick = false
			println("Match found: BeforeEntry equals entry.ID, kick set to false")

			break
		} else {
			BeforeEntry = entry.ID

			cacheMutex.Lock()
			cache[cacheKey] = BeforeEntry
			cacheMutex.Unlock()

			kick = true
			println("No match found: BeforeEntry updated to:", BeforeEntry)
			println("kick set to true")

			break
		}

	}

	UserInfo, _ := s.User(UserID)

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID, lang := SelectDB("channel_log_serverID", gmr.GuildID)

	if channel_log_serverID == 0 {
		return
	}
	trs := getTranslation(lang)
	if kick {

		code, err := generateCode(6)
		if err != nil {
			Error("Помилка генерації коду", err)
		}
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s "+"`%s`", trs.Kick, code),
			Description: fmt.Sprintf(
				">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"**%s: **"+"__***%s***__",
				trs.User, gmr.User.ID, trs.ID, gmr.User.ID, trs.Reason, Reason,
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: gmr.AvatarURL("256"),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    UserInfo.Username,
				IconURL: UserInfo.AvatarURL("256"),
			},
			Color:     0xc43737, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		message, _ := s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
		ActionType = fmt.Sprintf("[**%s**]"+"(https://discord.com/channels/%s/%s/%s)", trs.Kick, gmr.GuildID, message.ChannelID, message.ID)
	}
	embed := &discordgo.MessageEmbed{
		Title: trs.UserLeftGuild,
		Description: fmt.Sprintf(
			">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"%s",
			trs.User, gmr.User.ID, trs.ID, gmr.User.ID, ActionType,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: gmr.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserBanned(s *discordgo.Session, ban *discordgo.GuildBanAdd) {
	var ActionType string

	AuditLog, err := s.GuildAuditLog(ban.GuildID, "", "", 22, 1)
	if err != nil {
		Error("AUDIT", err)
		return
	}

	var UserID string
	var Reason string

	for _, entry := range AuditLog.AuditLogEntries {
		UserID = entry.UserID
		Reason = entry.Reason
		println("UserID:", entry.UserID)
		println("Reason:", entry.Reason)

	}

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID, lang := SelectDB("channel_log_serverID", ban.GuildID)
	if channel_log_serverID == 0 {
		return
	}

	trs := getTranslation(lang)
	UserInfo, _ := s.User(UserID)

	code, err := generateCode(6)
	if err != nil {
		Error("Помилка генерації коду", err)
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s "+"`%s`", trs.Ban, code),
		Description: fmt.Sprintf(
			">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"**%s: **"+"__***%s***__"+"\n",
			trs.User, ban.User.ID, trs.ID, ban.User.ID, trs.Reason, Reason,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: ban.User.AvatarURL("256"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    UserInfo.Username,
			IconURL: UserInfo.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)

	embed = &discordgo.MessageEmbed{
		Title: trs.UserLeftGuild,
		Description: fmt.Sprintf(
			">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"%s",
			trs.User, ban.User.ID, trs.ID, ban.User.ID, ActionType,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: ban.User.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserMuted(s *discordgo.Session, mute *discordgo.GuildMemberUpdate) {

	if mute.BeforeUpdate == nil && mute.CommunicationDisabledUntil == nil {
		return
	}

	channel_log_serverID, lang := SelectDB("channel_log_serverID", mute.GuildID)
	trs := getTranslation(lang)

	if mute.BeforeUpdate != nil {
		AuditLog, err := s.GuildAuditLog(mute.GuildID, "", "", 24, 1)
		if err != nil {
			log.Printf("AUDIT ERROR: %v", err)
			return
		}

		var UserID string

		currentTime := time.Now().UTC()
		stringTime := currentTime.Format(time.RFC3339)

		for _, entry := range AuditLog.AuditLogEntries {

			UserID = entry.UserID
			break
		}

		UserInfo, _ := s.User(UserID)

		embed := &discordgo.MessageEmbed{
			Title: trs.UnMute,
			Description: fmt.Sprintf(
				">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"\n"+"**%s: **"+"<t:%d:R>",
				trs.User, mute.User.ID, trs.ID, mute.User.ID, trs.TimeRemove, mute.BeforeUpdate.CommunicationDisabledUntil.Unix(),
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: mute.AvatarURL("256"),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    UserInfo.Username,
				IconURL: UserInfo.AvatarURL("256"),
			},
			Color:     0x5fc437, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
	}

	if mute.CommunicationDisabledUntil != nil {
		AuditLog, err := s.GuildAuditLog(mute.GuildID, "", "", 24, 1)
		if err != nil {
			log.Printf("AUDIT ERROR: %v", err)
			return
		}

		var UserID string
		var Reason string

		currentTime := time.Now().UTC()
		stringTime := currentTime.Format(time.RFC3339)

		/*hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		seconds := int(duration.Seconds()) % 60*/

		for _, entry := range AuditLog.AuditLogEntries {

			UserID = entry.UserID
			Reason = entry.Reason
			break
		}

		UserInfo, _ := s.User(UserID)

		code, err := generateCode(6)
		if err != nil {
			Error("Помилка генерації коду", err)
		}

		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s "+"`%s`", trs.Mute, code),
			Description: fmt.Sprintf(
				">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n"+"**%s: **"+"__***%s***__"+"\n"+"**%s: **"+"<t:%d:R>",
				trs.User, mute.User.ID, trs.ID, mute.User.ID, trs.Reason, Reason, trs.TimeRemove, mute.CommunicationDisabledUntil.Unix(),
			),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: mute.AvatarURL("256"),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    UserInfo.Username,
				IconURL: UserInfo.AvatarURL("256"),
			},
			Color:     0xc43737, // Колір (у форматі HEX)
			Timestamp: stringTime,
		}
		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
	} else {
		return
	}
}
func UserUnBanned(s *discordgo.Session, unban *discordgo.GuildBanRemove) {
	currentTime := time.Now().UTC()
	stringTime := currentTime.Format(time.RFC3339)

	AuditLog, err := s.GuildAuditLog(unban.GuildID, "", "", 23, 1)
	if err != nil {
		Error("AUDIT", err)
		return
	}

	var UserID string

	for _, entry := range AuditLog.AuditLogEntries {
		UserID = entry.UserID
	}

	UserInfo, _ := s.User(UserID)

	channel_log_serverID, lang := SelectDB("channel_log_serverID", unban.GuildID)

	trs := getTranslation(lang)

	embed := &discordgo.MessageEmbed{
		Title: trs.UnBan,
		Description: fmt.Sprintf(
			">>> **%s: **"+"<@%s>"+"\n"+"**%s: **"+"`%s`"+"\n",
			trs.User, unban.User.ID, trs.ID, unban.User.ID,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: unban.User.AvatarURL("256"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    UserInfo.Username,
			IconURL: UserInfo.AvatarURL("256"),
		},
		Color:     0x5fc437, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
