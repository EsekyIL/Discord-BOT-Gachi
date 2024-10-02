package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lmittmann/tint"
)

func ColumnExists(database *sql.DB, columnName string) (bool, error) {
	query := `SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_NAME = 'servers' AND COLUMN_NAME = ?`
	var count int
	err := database.QueryRow(query, columnName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error querying information schema: %v", err)
	}

	return count > 0, nil
}

func registerServer(sess *discordgo.Session, g *discordgo.GuildCreate) {
	database, err := ConnectDB() // Заміна на вашу функцію підключення
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	exists, err := ColumnExists(database, g.Guild.ID)
	if err != nil {
		log.Printf("Error checking column existence: %v", err)
		return
	}

	if !exists {
		return
	} else {
		owner, err := sess.User(g.OwnerID)
		if err != nil {
			Error("error query", err)
			return
		}
		query := `INSERT INTO servers (guild_id, guild_name, guild_owner_id, guild_owner_name, members, vip, forum, channel_id_forum, channel_id_message, channel_id_voice, channel_id_server, channel_id_penalty) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		err = UpdateDB(query, g.Guild.ID, g.Guild.Name, g.OwnerID, owner.GlobalName, g.MemberCount, 0, 0, 0, 0, 0, 0, 0)
		if err != nil {
			Error("error insert data", err)
			return
		}
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
