package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var cache = make(map[string]string)
var cacheMutex sync.RWMutex

func getCreationTime(userID string) time.Time {
	id, _ := strconv.ParseInt(userID, 10, 64)

	timestamp := (id >> 22) + 1420070400000

	return time.Unix(0, timestamp*int64(time.Millisecond))
}

func InvCreate(s *discordgo.Session, ic *discordgo.InviteCreate) {
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(ic.GuildID), ic.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
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
		Title: "An invitation has been created",
		Description: fmt.Sprintf(
			">>> **Code: **"+"`%s`"+"\n"+"**Channel: **"+"<#%s>"+"\n"+"**Validity period: **"+"<t:%d:R>"+"\n"+"**Count of users: **"+"%s",
			ic.Code, ic.ChannelID, expiresAt, MaxUses,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    ic.Inviter.Username,
			IconURL: ic.Inviter.AvatarURL("256"),
		},
		Color:     0x37c4b8, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("invite create problem", err)
		return
	}
}
func UserJoin(s *discordgo.Session, gma *discordgo.GuildMemberAdd) {
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(gma.GuildID), gma.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}

	userCreatedAt := getCreationTime(gma.User.ID)

	embed := &discordgo.MessageEmbed{
		Title: "New user",
		Description: fmt.Sprintf(
			">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"**Created: **"+"<t:%d:R>"+"\n",
			gma.User.ID, gma.User.ID, int64(userCreatedAt.Unix()),
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: gma.AvatarURL("256"),
		},
		Color:     0x5fc437, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error guild member add", err)
		return
	}
}
func UserExit(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) {
	if gmr.Member.User.ID == "1160175895475138611" {
		return
	}
	cacheKey := fmt.Sprintf("%s:%s", gmr.GuildID, gmr.User.ID)

	cacheMutex.RLock()
	BeforeEntry, exists := cache[cacheKey]
	cacheMutex.RUnlock()

	if !exists {
		BeforeEntry = ""
	}

	AuditLog, err := s.GuildAuditLog(gmr.GuildID, "", "", 20, 1)
	if err != nil {
		Error("error parsing Audit Log", err)
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
			// Match found: BeforeEntry equals entry.ID, kick set to false
			break
		} else {
			BeforeEntry = entry.ID
			cacheMutex.Lock()
			cache[cacheKey] = BeforeEntry
			cacheMutex.Unlock()
			kick = true
			// No match found: BeforeEntry updated. Kick set to true
			break
		}

	}
	UserInfo, _ := s.User(UserID)

	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(gmr.GuildID), gmr.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}

	if kick {

		code, err := generateCode(6)
		if err != nil {
			Error("error generation Code", err)
		}
		embed := &discordgo.MessageEmbed{
			Title: "Kick " + "`" + code + "`",
			Description: fmt.Sprintf(
				">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"**Reason: **"+"__***%s***__",
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
			Timestamp: time.Now().Format(time.RFC3339),
		}
		message, _ := s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
		ActionType = fmt.Sprintf("[**KICKED**]"+"(https://discord.com/channels/%s/%s/%s)", gmr.GuildID, message.ChannelID, message.ID)
	}
	embed := &discordgo.MessageEmbed{
		Title: "The user has left the guild",
		Description: fmt.Sprintf(
			">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"%s",
			gmr.User.ID, gmr.User.ID, ActionType,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: gmr.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error in leave people", err)
		return
	}
}
func UserBanned(s *discordgo.Session, ban *discordgo.GuildBanAdd) {
	var ActionType string

	AuditLog, err := s.GuildAuditLog(ban.GuildID, "", "", 22, 1)
	if err != nil {
		Error("error parsing Audit Log x2", err)
		return
	}

	var UserID string
	var Reason string

	for _, entry := range AuditLog.AuditLogEntries {
		UserID = entry.UserID
		Reason = entry.Reason
	}

	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(ban.GuildID), ban.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}

	UserInfo, _ := s.User(UserID)

	code, err := generateCode(6)
	if err != nil {
		Error("error generation Code", err)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Ban " + "`" + code + "`",
		Description: fmt.Sprintf(
			">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"**Reason: **"+"__***%s***__"+"\n",
			ban.User.ID, ban.User.ID, Reason,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: ban.User.AvatarURL("256"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    UserInfo.Username,
			IconURL: UserInfo.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error user banned!", err)
	}

	embed = &discordgo.MessageEmbed{
		Title: "The user has left the guild",
		Description: fmt.Sprintf(
			">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"%s",
			ban.User.ID, ban.User.ID, ActionType,
		),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: ban.User.AvatarURL("256"),
		},
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error user left guild x2", err)
		return
	}
}
func UserMuted(s *discordgo.Session, mute *discordgo.GuildMemberUpdate) {
	if mute.BeforeUpdate == nil && mute.CommunicationDisabledUntil == nil {
		return
	}

	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(mute.GuildID), mute.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}

	if mute.BeforeUpdate != nil {
		AuditLog, err := s.GuildAuditLog(mute.GuildID, "", "", 24, 1)
		if err != nil {
			Error("error parsing Audit Log x3", err)
			return
		}

		var UserID string

		for _, entry := range AuditLog.AuditLogEntries {

			UserID = entry.UserID
			break
		}

		UserInfo, _ := s.User(UserID)

		embed := &discordgo.MessageEmbed{
			Title: "Unmuted",
			Description: fmt.Sprintf(
				">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"\n"+"**Time to remove the restriction: **"+"<t:%d:R>",
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
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
		if err != nil {
			Error("error user unmute", err)
			return
		}
	}

	if mute.CommunicationDisabledUntil != nil {
		AuditLog, err := s.GuildAuditLog(mute.GuildID, "", "", 24, 1)
		if err != nil {
			Error("error parsing Audit Log x4", err)
			return
		}

		var UserID string
		var Reason string

		for _, entry := range AuditLog.AuditLogEntries {

			UserID = entry.UserID
			Reason = entry.Reason
			break
		}

		UserInfo, _ := s.User(UserID)

		code, err := generateCode(6)
		if err != nil {
			Error("error generation code x3", err)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: "Mute " + "`" + code + "`",
			Description: fmt.Sprintf(
				">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n"+"**Reason: **"+"__***%s***__"+"\n"+"**Time to remove the restriction: **"+"<t:%d:R>",
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
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
		if err != nil {
			Error("error mute member", err)
			return
		}
	} else {
		return
	}
}
func UserUnBanned(s *discordgo.Session, unban *discordgo.GuildBanRemove) {

	AuditLog, err := s.GuildAuditLog(unban.GuildID, "", "", 23, 1)
	if err != nil {
		Error("error parsing Audit Log x5 ", err)
		return
	}

	var UserID string

	for _, entry := range AuditLog.AuditLogEntries {
		UserID = entry.UserID
	}

	UserInfo, _ := s.User(UserID)

	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(unban.GuildID), unban.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Unban",
		Description: fmt.Sprintf(
			">>> **User: **"+"<@%s>"+"\n"+"**ID: **"+"`%s`"+"\n",
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
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error member unbanned", err)
		return
	}
}
func RoleCreated(s *discordgo.Session, rc *discordgo.GuildRoleCreate) {
	var roleName string
	var user *discordgo.User
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(rc.GuildID), rc.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}

	AuditLog, err := s.GuildAuditLog(rc.GuildID, "", "", 30, 1)
	if err != nil {
		Error("error parsing Audit Log x6 ", err)
		return
	}
	for _, entry := range AuditLog.AuditLogEntries {
		user, _ = s.User(entry.UserID)
	}

	if rc.Role.Name == "нова роль" || rc.Role.Name == "new role" {
		roleName = ""
	} else {
		roleName = rc.Role.Name
	}

	embed := &discordgo.MessageEmbed{
		Title: "Role create",
		Description: fmt.Sprintf(
			"Role %s successfully created",
			roleName,
		),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    user.Username,
			IconURL: user.AvatarURL("256"),
		},
		Color:     0x5fc437, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error role created", err)
		return
	}
}
func RoleDeleted(s *discordgo.Session, rd *discordgo.GuildRoleDelete) {
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(rd.GuildID), rd.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}
	var user *discordgo.User
	AuditLog, err := s.GuildAuditLog(rd.GuildID, "", "", 32, 1)
	if err != nil {
		Error("error parsing Audit Log x7 ", err)
		return
	}
	for _, entry := range AuditLog.AuditLogEntries {
		user, _ = s.User(entry.UserID)
	}
	role, err := s.State.Role(rd.GuildID, rd.RoleID)
	if err != nil {
		Error("error parsing info about role", err)
		embed := &discordgo.MessageEmbed{
			Title:       "Role delete",
			Description: "Role successfully deleted",
			Color:       0xc43737, // Колір (у форматі HEX)
			Timestamp:   time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text:    user.Username,
				IconURL: user.AvatarURL("256"),
			},
		}
		_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
		if err != nil {
			Error("error role deleted", err)
			return
		}
		return
	}
	embed := &discordgo.MessageEmbed{
		Title: "Role delete",
		Description: fmt.Sprintf(
			"Role %s successfully deleted",
			role.Name,
		),
		Color:     0xc43737, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error role deleted", err)
		return
	}
}
func RoleUpdated(s *discordgo.Session, ru *discordgo.GuildRoleUpdate) {
	var roleID string
	AuditLog, err := s.GuildAuditLog(ru.GuildID, "", "", 31, 1)
	time.Sleep(4 * time.Second)
	if err != nil {
		Error("error parsing Audit Log x8 ", err)
		return
	}
	for _, entry := range AuditLog.AuditLogEntries {
		roleID = entry.TargetID
	}
	if ru.Role.ID != roleID {
		return
	}
	rows, err := SelectDB(fmt.Sprintf("SELECT * FROM %s WHERE id = %s", shortenNumber(ru.GuildID), ru.GuildID))
	if err != nil {
		Error("error parsing data in DB", err)
	}
	if rows.Channel_ID_Server == "0" {
		return
	}
	embed := &discordgo.MessageEmbed{
		Title: "Role updated",
		Description: fmt.Sprintf(
			"Role `%s`, <@&%s> successfully updated",
			ru.Role.Name, ru.Role.ID,
		),
		Color:     0xc4b137, // Колір (у форматі HEX)
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbed(rows.Channel_ID_Server, embed)
	if err != nil {
		Error("error role updated", err)
		return
	}
}
