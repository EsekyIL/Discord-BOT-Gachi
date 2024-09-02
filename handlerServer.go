package main

import (
	"database/sql"
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
		Color:     0x37c4b8, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserJoin(s *discordgo.Session, gma *discordgo.GuildMemberAdd, database *sql.DB) {
	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID := SelectDB("channel_log_serverID", gma.GuildID, database)
	if channel_log_serverID == 0 {
		return
	}

	userCreatedAt := getCreationTime(gma.User.ID)

	embed := &discordgo.MessageEmbed{
		Title: "Новий користувач",
		Description: fmt.Sprintf(
			">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"**Створений: **"+"<t:%d:R>"+"\n",
			gma.User.ID, gma.User.ID, int64(userCreatedAt.Unix()),
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: gma.AvatarURL("256"),
		},
		Color:     0x5fc437, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserExit(s *discordgo.Session, gmr *discordgo.GuildMemberRemove, database *sql.DB) {

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

		println("UserID:", entry.UserID)
		println("Reason:", entry.Reason)
		println("Entry ID:", entry.ID)
		println("Current BeforeEntry:", BeforeEntry)
		println("------------------------")

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

	println("Final BeforeEntry:", BeforeEntry)
	println("Final kick value:", kick)

	UserInfo, _ := s.User(UserID)

	currentTime := time.Now()
	stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

	channel_log_serverID := SelectDB("channel_log_serverID", gmr.GuildID, database)
	if channel_log_serverID == 0 {
		return
	}
	if kick {
		channel_log_punishmentID := SelectDB("channel_log_punishmentID", gmr.GuildID, database)
		code, err := generateCode(6)
		if err != nil {
			Error("Помилка генерації коду", err)
		}
		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Кік "+"`%s`", code),
			Description: fmt.Sprintf(
				">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"**Причина: **"+"__***%s***__",
				gmr.User.ID, gmr.User.ID, Reason,
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
		message, _ := s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_punishmentID), embed)
		ActionType = fmt.Sprintf("[**Kicked**]"+"(https://discord.com/channels/%s/%s/%s)", gmr.GuildID, message.ChannelID, message.ID)
	}
	embed := &discordgo.MessageEmbed{
		Title: "Користувач покинув гільдію",
		Description: fmt.Sprintf(
			">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"%s",
			gmr.User.ID, gmr.User.ID, ActionType,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: gmr.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserBanned(s *discordgo.Session, ban *discordgo.GuildBanAdd, database *sql.DB) {
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

	channel_log_serverID := SelectDB("channel_log_serverID", ban.GuildID, database)
	if channel_log_serverID == 0 {
		return
	}

	UserInfo, _ := s.User(UserID)

	channel_log_punishmentID := SelectDB("channel_log_punishmentID", ban.GuildID, database)
	code, err := generateCode(6)
	if err != nil {
		Error("Помилка генерації коду", err)
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Бан "+"`%s`", code),
		Description: fmt.Sprintf(
			">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"**Причина: **"+"__***%s***__"+"\n",
			ban.User.ID, ban.User.ID, Reason,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: UserInfo.AvatarURL("256"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    UserInfo.Username,
			IconURL: UserInfo.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_punishmentID), embed)

	embed = &discordgo.MessageEmbed{
		Title: "Користувач покинув гільдію",
		Description: fmt.Sprintf(
			">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"%s",
			ban.User.ID, ban.User.ID, ActionType,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: ban.User.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: stringTime,
	}
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_serverID), embed)
}
func UserMuted(s *discordgo.Session, mute *discordgo.GuildMemberUpdate, database *sql.DB) {

	if mute.BeforeUpdate != nil {
		fmt.Printf("aeza %s", mute.BeforeUpdate.CommunicationDisabledUntil)
		AuditLog, err := s.GuildAuditLog(mute.GuildID, "", "", 24, 1)
		if err != nil {
			log.Printf("AUDIT ERROR: %v", err)
			return
		}

		var UserID string

		currentTime := time.Now().UTC()
		stringTime := currentTime.Format(time.RFC3339)

		/*hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		seconds := int(duration.Seconds()) % 60*/

		for _, entry := range AuditLog.AuditLogEntries {

			UserID = entry.UserID
			break
		}

		UserInfo, _ := s.User(UserID)

		channel_log_punishmentID := SelectDB("channel_log_punishmentID", mute.GuildID, database)

		embed := &discordgo.MessageEmbed{
			Title: "Размут",
			Description: fmt.Sprintf(
				">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"\n"+"**Час зняття обмеження: **"+"<t:%d:R>",
				mute.User.ID, mute.User.ID, mute.BeforeUpdate.CommunicationDisabledUntil.Unix(),
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
		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_punishmentID), embed)
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

		channel_log_punishmentID := SelectDB("channel_log_punishmentID", mute.GuildID, database)
		code, err := generateCode(6)
		if err != nil {
			Error("Помилка генерації коду", err)
		}

		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Мут "+"`%s`", code),
			Description: fmt.Sprintf(
				">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n"+"**Причина: **"+"__***%s***__"+"\n"+"**Час зняття обмеження: **"+"<t:%d:R>",
				mute.User.ID, mute.User.ID, Reason, mute.CommunicationDisabledUntil.Unix(),
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
		_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_punishmentID), embed)
	} else {
		return
	}
}
func UserUnBanned(s *discordgo.Session, unban *discordgo.GuildBanRemove, database *sql.DB) {
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

	channel_log_punishmentID := SelectDB("channel_log_punishmentID", unban.GuildID, database)

	embed := &discordgo.MessageEmbed{
		Title: "Разбан",
		Description: fmt.Sprintf(
			">>> **Користувач: **"+"<@%s>"+"\n"+"**Айді: **"+"`%s`"+"\n",
			unban.User.ID, unban.User.ID,
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
	_, _ = s.ChannelMessageSendEmbed(strconv.Itoa(channel_log_punishmentID), embed)
}
