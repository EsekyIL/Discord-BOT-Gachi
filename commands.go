package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ChannelsID []string

func ParseFlexibleDuration(input string) (int64, error) {
	// Регулярний вираз для парсингу часу у форматі 20m, 5h, 30s, 4d 20h 25m 10s і подібних
	re := regexp.MustCompile(`(\d+)([smhd])`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format")
	}

	var totalDuration time.Duration

	// Проходимо по всіх знайдених частинах тривалості (наприклад, 4d, 20h, 25m, 10s)
	for _, match := range matches {
		// Парсимо число
		amount, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %v", err)
		}

		// Визначаємо одиницю виміру часу (s = seconds, m = minutes, h = hours, d = days)
		unit := match[2]
		switch unit {
		case "s":
			totalDuration += time.Duration(amount) * time.Second
		case "m":
			totalDuration += time.Duration(amount) * time.Minute
		case "h":
			totalDuration += time.Duration(amount) * time.Hour
		case "d":
			totalDuration += time.Duration(amount) * time.Hour * 24 // Додаємо дні
		default:
			return 0, fmt.Errorf("unknown time unit: %s", unit)
		}
	}

	// Додаємо загальну тривалість до поточного часу і отримуємо UNIX timestamp
	unixTime := time.Now().Add(totalDuration).Unix()
	return unixTime, nil
}
func isAdmin(s *discordgo.Session, ic *discordgo.InteractionCreate) bool {
	member := ic.Interaction.Member.Roles

	for _, roleID := range member {
		role, err := s.State.Role(ic.Interaction.GuildID, roleID)
		if err != nil {
			Error("Error fetching role:", err)
			continue
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}
	return false
}
func registerCommands(sess *discordgo.Session) {

	selectMenu := discordgo.ApplicationCommand{
		Name:        "settings",
		Description: "select an option in settings",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err := sess.ApplicationCommandCreate("1160175895475138611", "", &selectMenu)
	if err != nil {
		Error("Error creating command settings", err)
		return
	}
	giveawayCreate := discordgo.ApplicationCommand{
		Name:        "gcreate",
		Description: "start giveaway (modal window)",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", &giveawayCreate)
	if err != nil {
		Error("Error creating command gcreate", err)
		return
	}

	ticketCreate := discordgo.ApplicationCommand{
		Name:        "ticket",
		Description: "create ticket in forum",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", &ticketCreate)
	if err != nil {
		Error("Error creating command ticket", err)
		return
	}
	reportCreate := discordgo.ApplicationCommand{
		Name:        "report",
		Description: "reate a report for the bot creator",
		Type:        discordgo.ChatApplicationCommand,
	}
	_, err = sess.ApplicationCommandCreate("1160175895475138611", "", &reportCreate)
	if err != nil {
		Error("Error creating command ticket", err)
		return
	}
	/*err = sess.ApplicationCommandDelete("1160175895475138611", "", "1177394341371707423") ITS FIND AND DELETE COMMANDS, do not erase!!!!!!!!!!
	if err != nil {
		Error("Error delete command", err)
		return
	}
	err = sess.ApplicationCommandDelete("1160175895475138611", "", "1177398356906082368")
	if err != nil {
		Error("Error delete command", err)
		return
	}
	comands, err := sess.ApplicationCommands("1160175895475138611", "")
	if err != nil {
		Error("Error parsing commands", err)
		return
	}
	for _, comand := range comands {
		println(comand.Name)
		println(comand.ID)
		println("=======================")

	}*/
}
func Commands(s *discordgo.Session, ic *discordgo.InteractionCreate) {

	if ic.Type == discordgo.InteractionApplicationCommand {
		switch ic.ApplicationCommandData().Name {
		case "report":
			if !isAdmin(s, ic) {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You do not have permission to use this command.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "report-create",
					Title:    "Create a report",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "guildID",
									Label:     "Guild ID",
									Style:     discordgo.TextInputShort,
									MinLength: 6,
									MaxLength: 24,
									Value:     ic.GuildID,
									Required:  true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "authorID",
									Label:     "author ID",
									Style:     discordgo.TextInputShort,
									Value:     ic.Member.User.ID,
									MinLength: 6,
									MaxLength: 24,
									Required:  true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "title",
									Label:       "title",
									Placeholder: "Enter a title...",
									Style:       discordgo.TextInputShort,
									MinLength:   4,
									MaxLength:   24,
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "description",
									Label:       "description",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Enter the description of report...",
									Required:    true,
									MinLength:   100,
									MaxLength:   2000,
								},
							},
						},
					},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Giveaway error creating", err)
			}
			return

		case "ticket":
			row, err := SelectDB(`SELECT * FROM servers WHERE guild_id = ?`, ic.GuildID)
			if err != nil {
				Error("error parsing data in DB", err)
			}
			if !row.forum {
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "ticket-create",
					Title:    "Create a ticket",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "forumID",
									Label:     "forum id",
									Style:     discordgo.TextInputShort,
									Value:     row.channel_id_forum,
									Required:  true,
									MinLength: 1,
									MaxLength: 23,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "author",
									Label:     "author id",
									Value:     ic.Member.User.ID,
									Style:     discordgo.TextInputShort,
									Required:  true,
									MinLength: 3,
									MaxLength: 23,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "title",
									Label:     "Title",
									Style:     discordgo.TextInputShort,
									Required:  true,
									MinLength: 3,
									MaxLength: 50,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "tag",
									Label:       "Tag",
									Placeholder: "0 - not use, 1 - first, 2 - second, 3 - all",
									Style:       discordgo.TextInputShort,
									Required:    true,
									MinLength:   1,
									MaxLength:   1,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "description",
									Label:       "description",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Enter the description of the ticket...",
									Required:    true,
									MinLength:   0,
									MaxLength:   1000,
								},
							},
						},
					},
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Giveaway error creating", err)
			}
			return
		case "gcreate":
			if !isAdmin(s, ic) {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You do not have permission to use this command.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "giveaway-create",
					Title:    "Create a giveaway",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "winners",
									Label:     "number of winners",
									Style:     discordgo.TextInputShort,
									MinLength: 1,
									MaxLength: 3,
									Value:     "1",
									Required:  true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "duration",
									Label:       "duration",
									Style:       discordgo.TextInputShort,
									Placeholder: "Ex: 20d 35h 20m",
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "prize",
									Label:       "prize",
									Placeholder: "Enter a title...",
									Style:       discordgo.TextInputShort,
									Required:    true,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "description",
									Label:       "description",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Enter the description of the giveaway...",
									Required:    false,
									MinLength:   0,
									MaxLength:   1000,
								},
							},
						},
					},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Giveaway error creating", err)
			}

			return

		case "settings":
			if !isAdmin(s, ic) {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You do not have permission to use this command.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("", err)
				}
				return
			}
			embed := &discordgo.MessageEmbed{
				Title: "Setting up bot functions",
				Description: "This bot helps you maintain a detailed log of server events, automatically tracking important user actions. " +
					"The ticket system allows members to easily create support or issue requests, making it simple to manage and monitor everything in one place. \n\n" +
					"For more information, you can click the button to access articles about the bot.",
				Footer: &discordgo.MessageEmbedFooter{
					Text:    ic.Interaction.Member.User.Username,
					IconURL: ic.Interaction.Member.User.AvatarURL("256"),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			}

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Logging",
									Style:    discordgo.PrimaryButton,
									Disabled: false,
									Emoji: &discordgo.ComponentEmoji{
										Name: "📋",
									},
									CustomID: "logging-btn",
								},
								discordgo.Button{
									Label:    "Ticket system",
									Style:    discordgo.PrimaryButton,
									Disabled: false,
									Emoji: &discordgo.ComponentEmoji{
										Name: "🎫",
									},
									CustomID: "ticket-btn",
								},
								discordgo.Button{
									Label:    "About",
									Style:    discordgo.LinkButton,
									Disabled: false,
									Emoji: &discordgo.ComponentEmoji{
										Name: "🔗",
									},
									URL: "https://google.com",
								},
							},
						},
					},
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("error send command settings", err)
			}
		}
		return
	} else if ic.Type == discordgo.InteractionMessageComponent {
		switch ic.MessageComponentData().CustomID {
		case "logging-btn":
			minValues := 4
			embed := &discordgo.MessageEmbed{
				Title: "Logging",
				Description: "Click the button to automatically create login channels, or select them manually. \n\n" +
					"***Or select the channels in this order:***" + "\n" +
					"1. Message channel\n" +
					"2. Voice chat channel\n" +
					"3. Event channel on the server\n" +
					"4. Channel of punishments\n",
				Footer: &discordgo.MessageEmbedFooter{
					Text:    ic.Interaction.Member.User.Username,
					IconURL: ic.Interaction.Member.User.AvatarURL("256"),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			}
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							MenuType:     discordgo.ChannelSelectMenu,
							MinValues:    &minValues,
							MaxValues:    4,
							CustomID:     "channel-slct",
							Placeholder:  "It is necessary to poke here",
							ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Auto",
							Style:    discordgo.SuccessButton,
							Disabled: false,
							Emoji: &discordgo.ComponentEmoji{
								Name: "📋",
							},
							CustomID: "auto-btn",
						},
					},
				},
			}
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel:    ic.ChannelID,
				ID:         ic.Message.ID,
				Embed:      embed,
				Components: &components,
			})
			if err != nil {
				Error("Channel message edit complex", err)
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond", err)
			}
			return
		case "channel-slct":
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Now you can use the logging function",
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond in participates ", err)
			}
			channelsID := ic.MessageComponentData().Values

			query := `UPDATE servers SET channel_id_message = ?, channel_id_voice = ?, channel_id_server = ?, channel_id_penalty = ? WHERE guild_id = ?`
			go UpdateDB(query, channelsID[0], channelsID[1], channelsID[2], channelsID[3], ic.GuildID)
			s.ChannelMessageDelete(ic.ChannelID, ic.Message.ID)
			return

		case "auto-btn":
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Now you can use the logging function",
				},
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond in participates ", err)
			}
			category, err := s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "log",
				Type:     discordgo.ChannelTypeGuildCategory,
				Position: 1,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:   ic.GuildID,
						Type: discordgo.PermissionOverwriteTypeRole,
						Deny: discordgo.PermissionViewChannel,
					},
				},
			})
			if err != nil {
				Error("error guild channel category create", err)
			}
			channel_server, err := s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "server",
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: category.ID,
				Position: 3,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:   ic.GuildID,
						Type: discordgo.PermissionOverwriteTypeRole,
						Deny: discordgo.PermissionViewChannel,
					},
				},
			})
			if err != nil {
				fmt.Println("Помилка при створенні текстового каналу 1,", err)
				return
			}
			channel_msg, err := s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "message",
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: category.ID,
				Position: 1,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:   ic.GuildID,
						Type: discordgo.PermissionOverwriteTypeRole,
						Deny: discordgo.PermissionViewChannel,
					},
				},
			})
			if err != nil {
				fmt.Println("Помилка при створенні текстового каналу 1,", err)
				return
			}
			channel_voice, err := s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "voice",
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: category.ID,
				Position: 2,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:   ic.GuildID,
						Type: discordgo.PermissionOverwriteTypeRole,
						Deny: discordgo.PermissionViewChannel,
					},
				},
			})
			if err != nil {
				fmt.Println("Помилка при створенні текстового каналу 1,", err)
				return
			}
			channel_penalty, err := s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "penalty",
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: category.ID,
				Position: 4,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:   ic.GuildID,
						Type: discordgo.PermissionOverwriteTypeRole,
						Deny: discordgo.PermissionViewChannel,
					},
				},
			})
			if err != nil {
				fmt.Println("Помилка при створенні текстового каналу 1,", err)
				return
			}
			query := `UPDATE servers SET channel_id_message = ?, channel_id_voice = ?, channel_id_server = ?, channel_id_penalty = ? WHERE guild_id = ?`
			go UpdateDB(query, channel_msg.ID, channel_voice.ID, channel_server.ID, channel_penalty.ID, ic.GuildID)
			s.ChannelMessageDelete(ic.ChannelID, ic.Message.ID)
			return
		case "ticket-btn":
			embed := &discordgo.MessageEmbed{
				Title: "About ticket system",
				Description: "The ticket system is a feedback system with the `administration/moderation` of the guild. " +
					"The bot will automatically create a forum channel. Then the `/ticket` command will be available. " +
					"Configure access rights yourself!\n\n" + "***__Recommended: disable the right to create posts for all users.__***",
			}
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style:    discordgo.SuccessButton,
							Disabled: false,
							CustomID: "start-ticket",
							Emoji: &discordgo.ComponentEmoji{
								Name: "🚀",
							},
						},
					},
				},
			}
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel:    ic.ChannelID,
				ID:         ic.Message.ID,
				Embed:      embed,
				Components: &components,
			})
			if err != nil {
				Error("Channel message edit complex", err)
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond", err)
			}
			return
		case "start-ticket":
			s.ChannelMessageDelete(ic.ChannelID, ic.Message.ID)
			var everyoneID string
			roles, err := s.GuildRoles(ic.GuildID)
			if err != nil {
				Error("parsing roles", err)
				return
			}
			for _, role := range roles {
				if role.Name == "@everyone" {
					everyoneID = role.ID
					break
				}
			}
			forum, err := s.GuildChannelCreateComplex(ic.GuildID, discordgo.GuildChannelCreateData{
				Name:     "tickets",
				Type:     discordgo.ChannelTypeGuildForum,
				Topic:    "The bot automatically creates posts, no intervention is required.",
				Position: 1,
				NSFW:     false,
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					{
						ID:    everyoneID,
						Type:  discordgo.PermissionOverwriteTypeRole,
						Allow: discordgo.PermissionViewChannel | discordgo.PermissionReadMessageHistory | discordgo.PermissionSendMessagesInThreads,
						Deny:  discordgo.PermissionAll,
					},
				},
			})

			if err != nil {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: "**Error 😔**" + "\nTo create a forum, you need to make the server a community!",
					},
				}
				err = s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Interaction respond create forum-channel", err)
				}
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "*Success!!! 🙂*",
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond create forum-channel", err)
			}
			_, err = s.ChannelEditComplex(forum.ID, &discordgo.ChannelEdit{
				AvailableTags: &[]discordgo.ForumTag{
					{
						Name:      "Report",
						Moderated: false,
						EmojiName: "📢",
					},
					{
						Name:      "Question",
						Moderated: false,
						EmojiName: "❔",
					},
				},
			})
			if err != nil {
				Error("error create tags!", err)
			}

			query := `UPDATE servers SET channel_id_forum = ?, forum = 1 WHERE guild_id = ?`
			go UpdateDB(query, forum.ID, ic.GuildID)
			return
		case "leave-giveaway":
			_, err := leaveUserGiveaway(ic.GuildID, ic.Interaction.Member.User.ID)
			if err != nil {
				Error("leave user in giveaway", err)
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: ">>> *You have successfully left the giveaway! 😔*",
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond in participates ", err)
			}
			return

		case "participate":
			currentTime := time.Now().Format(time.RFC3339)

			gvw, Participates, err := incrementParticipantCount(ic.GuildID, ic.Interaction.Member.User.ID, ic.Message.ID)
			if err != nil {
				Error("incrementParticipantCount", err)
				return
			}
			if Participates {
				response := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:   discordgo.MessageFlagsEphemeral,
						Content: ">>> *You have already entered this giveaway!*",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Style:    discordgo.DangerButton,
										Disabled: false,
										CustomID: "leave-giveaway",
										Emoji: &discordgo.ComponentEmoji{
											Name: "🔚",
										},
									},
								},
							},
						},
					},
				}
				err := s.InteractionRespond(ic.Interaction, response)
				if err != nil {
					Error("Interaction respond in participates ", err)
				}
				return
			}
			embed := &discordgo.MessageEmbed{
				Title: gvw.Title,
				Color: 0xfadb84,
				Description: fmt.Sprintf(gvw.Description+"\n\n"+">>> **Ends: **"+"<t:%d:R>"+"  "+"<t:%d:f>"+"\n"+"** Hosted by: **"+"<@%s>"+"\n"+"**Entries: **"+"`%d`"+"\n"+"**Winners: **"+"`%s`",
					gvw.TimeUnix, gvw.TimeUnix, ic.Interaction.Member.User.ID, gvw.CountParticipate, gvw.Winers,
				),
				Timestamp: currentTime,
			}
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Style:    discordgo.PrimaryButton,
							Disabled: false,
							CustomID: "participate",
							Emoji: &discordgo.ComponentEmoji{
								Name: "🎆",
							},
						},
					},
				},
			}
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel:    ic.ChannelID,
				ID:         ic.Message.ID,
				Embed:      embed,
				Components: &components, // Тут передаємо слайс без вказівника
			})
			if err != nil {
				Error("Channel message edit complex", err)
				return
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond", err)
			}
			return
		}
	} else if ic.Type == discordgo.InteractionModalSubmit {
		switch ic.ModalSubmitData().CustomID {
		case "report-create":
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond", err)
			}

			guildID := ic.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			authorID := ic.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			title := ic.ModalSubmitData().Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description := ic.ModalSubmitData().Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			author, _ := s.User(authorID)
			guild, _ := s.Guild(guildID)

			query := `INSERT INTO reports (guild_name, guild_id, author_name, author_id, title, description, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
			err = UpdateDB(query, guild.Name, guildID, author.GlobalName, authorID, title, description, time.Now().Format(time.RFC3339))
			if err != nil {
				log.Println("Error inserting record:", err)
			}

			return

		case "ticket-create":
			index := 0
			tags := []string{}

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			}
			err := s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("Interaction respond", err)
			}

			forumID := ic.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			authorID := ic.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			title := ic.ModalSubmitData().Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			tagUse := ic.ModalSubmitData().Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description := ic.ModalSubmitData().Components[4].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			author, err := s.User(authorID)
			if err != nil {
				Error("error parsing author info", err)
			}
			forum, err := s.Channel(forumID)
			if err != nil {
				Error("error parsing forum info", err)
			}
			switch tagUse {
			case "1":
				for _, tag := range forum.AvailableTags {
					tags = append(tags, tag.ID)
					break
				}
			case "2":
				for _, tag := range forum.AvailableTags {
					if index == 1 {
						tags = append(tags, tag.ID)
						break
					}
					index++
				}
			case "3":
				for _, tag := range forum.AvailableTags {
					tags = append(tags, tag.ID)
				}
			}

			_, err = s.ForumThreadStartComplex(forumID, &discordgo.ThreadStart{
				Name:                title,
				AutoArchiveDuration: 1440,
				Type:                discordgo.ChannelTypeGuildForum,
				Invitable:           false,
				RateLimitPerUser:    30,
				AppliedTags:         tags,
			},
				&discordgo.MessageSend{
					Content: description,
					Embed: &discordgo.MessageEmbed{
						Timestamp: time.Now().Format(time.RFC3339),
						Footer: &discordgo.MessageEmbedFooter{
							Text:    author.Username,
							IconURL: author.AvatarURL("256"),
						},
					},
				})
			if err != nil {
				fmt.Println("Error creating thread:", err)
				return
			}
			return
		case "giveaway-create":
			currentTime := time.Now().Format(time.RFC3339)

			CountParticipate := 0

			var description string
			description = ""

			winners := ic.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			duration := ic.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			prize := ic.ModalSubmitData().Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description = ic.ModalSubmitData().Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

			unixTime, err := ParseFlexibleDuration(duration)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: prize,
				Color: 0xfadb84,
				Description: fmt.Sprintf(description+"\n\n"+">>> **Ends: **"+"<t:%d:R>"+"  "+"<t:%d:f>"+"\n"+"** Hosted by: **"+"<@%s>"+"\n"+"**Entries: **"+"`%d`"+"\n"+"**Winners: **"+"`%s`",
					unixTime, unixTime, ic.Interaction.Member.User.ID, CountParticipate, winners,
				),
				Footer: &discordgo.MessageEmbedFooter{
					Text:    ic.Interaction.Member.User.Username,
					IconURL: ic.Interaction.Member.User.AvatarURL("256"),
				},
				Timestamp: currentTime,
			}
			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Style:    discordgo.PrimaryButton,
									Disabled: false,
									CustomID: "participate",
									Emoji: &discordgo.ComponentEmoji{
										Name: "🎆",
									},
								},
							},
						},
					},
				},
			}
			err = s.InteractionRespond(ic.Interaction, response)
			if err != nil {
				Error("", err)
			}

			go GiveawayCreated(ic.GuildID, prize, description, unixTime, CountParticipate, int(time.Now().Unix()), winners, ic.ChannelID, ic.Interaction.Member.User.ID)

		}
		return
	}
}
