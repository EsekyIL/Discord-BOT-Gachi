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

func registerServer(sess *discordgo.Session, g *discordgo.GuildCreate, database *sql.DB) {
	// Формування SQL-запиту для створення таблиці
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id BIGINT PRIMARY KEY, 
    name VARCHAR(60), 
    members INTEGER,
	owner VARCHAR(60), 
    vip tinyint(1) DEFAULT 0, 
	forum tinyint(1) DEFAULT 0, 
    channel_id_forum VARCHAR(30) DEFAULT '0', 
    channel_id_message VARCHAR(30) DEFAULT '0', 
    channel_id_voice VARCHAR(30) DEFAULT '0', 
	channel_id_server VARCHAR(30) DEFAULT '0',
	channel_id_penalty VARCHAR(30) DEFAULT '0'
)`, shortenNumber(g.Guild.ID))
	_, err := database.Exec(query)
	if err != nil {
		Error("error creating table", err)
	}
	owner, err := sess.User(g.OwnerID)
	if err != nil {
		Error("error query", err)
		return
	}
	// Формування SQL-запиту для вставки даних
	query = fmt.Sprintf(`INSERT INTO %s (id, name, members, owner) VALUES (?, ?, ?, ?)`, shortenNumber(g.Guild.ID))

	// Підготовка запиту для вставки даних
	statement, err := database.Prepare(query)
	if err != nil {
		Error("error query", err)
		return
	}
	defer statement.Close()

	// Виконання запиту на вставку даних
	_, err = statement.Exec(g.Guild.ID, g.Guild.Name, g.Guild.MemberCount, fmt.Sprintf("%s | %s", owner.Username, owner.ID))
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
