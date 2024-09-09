package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lmittmann/tint"
)

func registerServer(g *discordgo.GuildCreate, database *sql.DB) {
	// Формування SQL-запиту для створення таблиці
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id BIGINT PRIMARY KEY, 
    name VARCHAR(255), 
    members INTEGER, 
    channel_log_msgID VARCHAR(255), 
    channel_log_voiceID VARCHAR(255), 
    channel_log_serverID VARCHAR(255), 
    channel_log_punishmentID VARCHAR(255), 
    Language VARCHAR(10)
)`, shortenNumber(g.Guild.ID))
	_, err := database.Exec(query)
	if err != nil {
		Error("error creating table", err)
	}

	// Формування SQL-запиту для вставки даних
	query = fmt.Sprintf(`INSERT INTO %s (id, name, members, channel_log_msgID, channel_log_voiceID, channel_log_serverID, channel_log_punishmentID, Language) VALUES (?,?, ?, ?, ?, ?, ?, ?)`, shortenNumber(g.Guild.ID))

	// Підготовка запиту для вставки даних
	statement, err := database.Prepare(query)
	if err != nil {
		Error("error query", err)
		return
	}
	defer statement.Close()

	// Виконання запиту на вставку даних
	_, err = statement.Exec(g.Guild.ID, g.Guild.Name, g.Guild.MemberCount, 0, 0, 0, 0, "EU")
	if err != nil {
		Error("error executing INSERT statement", err)
		return
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
