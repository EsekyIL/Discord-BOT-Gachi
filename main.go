package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

func goDotEnvVariable(key string) string {

	// завантажити файл .env
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Помилка завантаження файлу .env")
	}
	return os.Getenv(key)
}
func Error(msg string, err error) {
	logger := slog.New(tint.NewHandler(os.Stderr, nil))

	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
	logger.Error(msg, "Помилка", err)
}

func main() {
	database, err := sql.Open("sqlite", "./gopher.db")
	if err != nil {
		Error("Failed to open the database", err)
	}
	defer database.Close()

	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY, name TEXT, members INTEGER, channel_log_msgID INTEGER, channel_log_voiceID INTEGER, channel_log_serverID Integer)")

	defer statement.Close()

	_, err = statement.Exec()
	if err != nil {
		Error("Failed to execute the SQL statement", err)
	}

	token := goDotEnvVariable("API_KEY")
	userChannels := make(map[string]string)
	userTimeJoinVoice := make(map[string]string)
	sess, _ := discordgo.New("Bot " + token) // Відкриття сессії з ботом

	registerCommands(sess)

	sess.StateEnabled = true
	sess.State.MaxMessageCount = 1000

	sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		rows, _ := database.Query("SELECT id, name, members FROM servers WHERE id = ?", g.Guild.ID)
		var id int
		var name string
		var members int

		if rows.Next() {
			err = rows.Scan(&id, &name, &members)
			if err != nil {
				Error("Failed to scan the row", err)
			}
		} else {
			if err := rows.Err(); err != nil {
				Error("Failed during iteration over rows", err)
			}
			go registerServer(g, database) // Виклик функції для реєстрації сервера, якщо дані не знайдено
			return
		}
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) { // Модуль логування оновленого повідомлення
		if m.Author == nil || m.Author.Bot {
			return
		}
		go MsgUpdate(s, m, database)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) { // Модуль логування видаленого повідомлення
		go MessageDeleteLog(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) { // Модуль логування входу/переходу/виходу в голосових каналах
		go VoiceLog(s, vs, &userChannels, &userTimeJoinVoice)
	})
	sess.AddHandler(func(s *discordgo.Session, gma *discordgo.GuildMemberAdd) { // Модуль логування надходження користувачів на сервер
		go InvitePeopleToServer(s, gma)
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) { // Модуль логування виходу користувачів з серверу
		go ExitPeopleFromServer(s, gmr)
	})
	sess.AddHandler(func(s *discordgo.Session, b *discordgo.GuildBanAdd) { // Модуль логування бану користувачів на сервер
		// Створення об'єкта для представлення забаненого користувача
		ba, err := s.GuildBan(b.GuildID, b.User.ID)
		if err != nil {
			// Обробка помилки, якщо не вдається отримати інформацію про забаненого користувача
			fmt.Println("Помилка при отриманні даних про забаненого користувача:", err)
			return
		}

		// Виклик функції для обробки події бану користувача
		go BanUserToServer(s, b, ba)
	})
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers // Доп. дозволи

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("The bot is online!")

	sc := make(chan os.Signal, 1) // Вимкнення бота CTRL+C
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
