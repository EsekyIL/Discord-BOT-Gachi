package main

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lmittmann/tint"
)

func registerServer(g *discordgo.GuildCreate, database *sql.DB) {
	statement, err := database.Prepare("INSERT INTO servers (id, name, members, channel_log_msgID, channel_log_voiceID, channel_log_serverID, channel_log_punishmentID, BeforeEntryID ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		Error("Проблемс", err)
	}
	defer statement.Close()
	_, err = statement.Exec(g.Guild.ID, g.Guild.Name, g.Guild.MemberCount, 0, 0, 0, 0, "")
	if err != nil {
		Error("Error executing statement", err)
	}

	logger := slog.New(tint.NewHandler(os.Stderr, nil))

	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	logger.Info("🎉 Урааа. Бота добавили на сервер!", "ім'я серверу", g.Guild.Name, "к-сть людей", g.Guild.MemberCount)
}
