package main

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var usedCodes = make(map[string]bool) // Зберігаємо всі згенеровані коди

type RowData struct {
	ID                 string
	Name               string
	Members            string
	Owner              string
	Vip                bool
	Forum              bool
	Channel_ID_Forum   string
	Channel_ID_Message string
	Channel_ID_Voice   string
	Channel_ID_Server  string
	Channel_ID_Penalty string
}

func generateCode(length int) (string, error) {
	for {
		code := make([]byte, length)
		for i := range code {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[i] = charset[num.Int64()]
		}

		codeStr := string(code)

		// Перевіряємо, чи вже існує такий код
		if _, exists := usedCodes[codeStr]; !exists {
			// Якщо ні, то додаємо його до мапи і повертаємо
			usedCodes[codeStr] = true
			return codeStr, nil
		}
		// Якщо код вже існує, генеруємо новий
	}
}
func shortenNumber(number string) string {
	hasher := md5.New()
	hasher.Write([]byte(number))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)[:8] // Обрізаємо до 8 символів
}
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

func SelectDB(query string) (*RowData, error) {
	var result RowData
	database, err := ConnectDB()
	if err != nil {
		return &result, err
	}

	defer database.Close()

	rows, err := database.Query(query)
	if err != nil {
		return &result, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Members,
			&result.Owner,
			&result.Vip,
			&result.Forum,
			&result.Channel_ID_Forum,
			&result.Channel_ID_Message,
			&result.Channel_ID_Voice,
			&result.Channel_ID_Server,
			&result.Channel_ID_Penalty,
		)
		if err != nil {
			return &result, err
		}
	}

	if err = rows.Err(); err != nil {
		return &result, err
	}

	return &result, nil
}
func UpdateDB(query string) error {

	database, err := ConnectDB()
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}
	defer database.Close()

	statement, err := database.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer statement.Close()

	_, err = statement.Exec()
	if err != nil {
		return fmt.Errorf("error executing query: %v", err)
	}

	return nil
}
func ConnectDB() (*sql.DB, error) {
	// Отримуємо DSN з тайм-аутом підключення на 10 секунд
	dsn := goDotEnvVariable("DSN") + "?timeout=10s"

	// Підключаємося до бази даних
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	// Перевіряємо, чи вдалося підключитися до бази
	err = database.Ping()
	if err != nil {
		database.Close() // Закриваємо підключення, якщо Ping не пройшов
		return nil, fmt.Errorf("unable to reach database: %v", err)
	}

	// Встановлюємо максимальний час простою та відкритих з'єднань
	database.SetConnMaxIdleTime(5 * time.Minute)
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)

	return database, nil
}

func main() {

	token := goDotEnvVariable("API_KEY")
	sess, _ := discordgo.New("Bot " + token)

	registerCommands(sess)

	sess.StateEnabled = true
	sess.State.MaxMessageCount = 1000
	sess.State.TrackVoice = true
	sess.State.TrackMembers = true
	sess.State.TrackRoles = true

	sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {

		database, err := ConnectDB()
		if err != nil {
			log.Fatal(err)
		}

		query := fmt.Sprintf("SHOW TABLES LIKE '%s'", shortenNumber(g.Guild.ID))
		rows, err := database.Query(query)
		if err != nil {
			fmt.Println("Error executing query:", err)
			return
		}
		defer rows.Close()

		// Перевіряємо, чи таблиця існує
		if rows.Next() {
			return
		}
		go registerServer(s, g, database) // Виклик функції для реєстрації сервера, якщо дані не знайдено

	})
	sess.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) {

		Commands(s, ic)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		go MsgUpdate(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		go MsgDelete(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		go VoiceLog(s, vs)
	})
	sess.AddHandler(func(s *discordgo.Session, ic *discordgo.InviteCreate) {
		go InvCreate(s, ic)
	})
	sess.AddHandler(func(s *discordgo.Session, gma *discordgo.GuildMemberAdd) {
		go UserJoin(s, gma)
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) {
		go UserExit(s, gmr)
	})
	sess.AddHandler(func(s *discordgo.Session, ban *discordgo.GuildBanAdd) {
		go UserBanned(s, ban)
	})
	sess.AddHandler(func(s *discordgo.Session, mute *discordgo.GuildMemberUpdate) {
		if mute.BeforeUpdate == nil || mute.CommunicationDisabledUntil == nil {
			return
		}
		go UserMuted(s, mute)
	})
	sess.AddHandler(func(s *discordgo.Session, unban *discordgo.GuildBanRemove) {
		go UserUnBanned(s, unban)
	})
	sess.AddHandler(func(s *discordgo.Session, rc *discordgo.GuildRoleCreate) {
		go RoleCreated(s, rc)
	})
	sess.AddHandler(func(s *discordgo.Session, rd *discordgo.GuildRoleDelete) {
		go RoleDeleted(s, rd)
	})
	sess.AddHandler(func(s *discordgo.Session, ru *discordgo.GuildRoleUpdate) {
		go RoleUpdated(s, ru)
	})

	go func() {
		// Запускаємо перевірку
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			checkGiveaways(sess)
		}
	}()

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers // Доп. дозволи

	err := sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	currentTime := time.Now()
	fmt.Println(currentTime.Format(time.RFC1123), "The bot is online!")

	sc := make(chan os.Signal, 1) // Вимкнення бота CTRL+C
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
