package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var usedCodes = make(map[string]bool) // Зберігаємо всі згенеровані коди

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

func SelectDB(select_row string, GuildID string, database *sql.DB) int {
	query := fmt.Sprintf("SELECT id, %s FROM servers WHERE id = ?", select_row)
	rows, err := database.Query(query, GuildID)
	if err != nil {
		Error("Щось сталось", err)
		return 0 // Повертаємо 0 у разі помилки
	}
	defer rows.Close()

	var id int
	var channel_log int

	if rows.Next() {
		err := rows.Scan(&id, &channel_log)
		if err != nil {
			Error("Failed to scan the row", err)
			return 0
		}
	} else {
		if err := rows.Err(); err != nil {
			Error("Failed during iteration over rows", err)
		}
		return 0
	}

	return channel_log
}

func main() {
	logger := slog.New(tint.NewHandler(os.Stderr, nil))
	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		}),
	))

	database, err := sql.Open("sqlite", "./gopher.db")
	if err != nil {
		Error("Failed to open the database", err)
	}
	defer database.Close()

	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY, name TEXT, members INTEGER, channel_log_msgID INTEGER, channel_log_voiceID INTEGER, channel_log_serverID INTEGER, channel_log_punishmentID INTEGER, BeforeEntryID TEXT)")
	defer statement.Close()
	_, err = statement.Exec()
	if err != nil {
		Error("Failed to execute the SQL statement", err)
	}

	token := goDotEnvVariable("API_KEY")
	sess, _ := discordgo.New("Bot " + token)

	go registerCommands(sess, database)

	sess.StateEnabled = true
	sess.State.MaxMessageCount = 1000
	sess.State.TrackVoice = true
	sess.State.TrackMembers = true
	sess.State.TrackRoles = true

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
	/*sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) { // Модуль логування оновленого повідомлення
		if m.Author == nil || m.Author.Bot {
			return
		}
		if m.Author.ID == "733375879480082526" {
			currentTime := time.Now()
			stringTime := currentTime.Format("2006-01-02T15:04:05.999Z07:00")

			embed := &discordgo.MessageEmbed{
				Title: "ВНИМАНИЕ АКЦИЯ!",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name: "**Ссылки**",
						Value: "https://t.me/ShitCoinTap_Bot/Game?startapp=QAoqko6j66D2_XpsDXG9mA" + "\n" + "https://t.me/hamsTer_kombat_bot/start?startapp=kentId6271166370" + "\n" + "https://t.me/TimeTONbot?start=Esekyil" + "\n" +
							"https://t.me/drumtap_bot?start=1716917988423593" + "\n" + "https://t.me/blum/app?startapp=ref_ir3PwjIYrt" + "\n" + "https://t.me/wcoin_tapbot?start=NjI3MTE2NjM3MA" + "\n" + "https://t.me/memefi_coin_bot?start=r_74176e6a6f" + "\n" +
							"https://t.me/cexio_tap_bot?start=1716557174796979",
						Inline: false,
					},
				},
				Description: fmt.Sprintf(
					"### Приветствуем! У нас стартовала уникальная акция для самых активных участников.\n" +
						"``` Ваша задача проста – выполнить все задания, а именно: перейти по инвайт-ссылкам в Telegram." +
						"Места ограничены (всего 10 участников). За выполнение всех условий вы получите эксклюзивную кастомную роль на нашем Discord-сервере." +
						"Успейте занять своё место и свяжитесь со мной в личных сообщениях (Discord), чтобы обсудить награду! ```",
				),
				Footer: &discordgo.MessageEmbedFooter{
					Text:    m.Author.Username,
					IconURL: m.Author.AvatarURL("256"), // URL для іконки (може бути порожнім рядком)
				},
				Color:     0x37c4b8, // Колір (у форматі HEX)
				Timestamp: stringTime,
			}

			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
		}
	})*/
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		go MsgUpdate(s, m, database)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		go MsgDelete(s, m, database)
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
		go VoiceLog(s, vs, database)
	})
	sess.AddHandler(func(s *discordgo.Session, ic *discordgo.InviteCreate) {
		go InvCreate(s, ic, database)
	})
	sess.AddHandler(func(s *discordgo.Session, gma *discordgo.GuildMemberAdd) {
		go UserJoin(s, gma, database)
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) {
		go UserExit(s, gmr, database)
	})
	sess.AddHandler(func(s *discordgo.Session, ban *discordgo.GuildBanAdd) {
		go UserBanned(s, ban, database)
	})
	sess.AddHandler(func(s *discordgo.Session, mute *discordgo.GuildMemberUpdate) {
		// Отримання аудиторських записів для сервера
		go UserMuted(s, mute, database)
	})
	sess.AddHandler(func(s *discordgo.Session, unban *discordgo.GuildBanRemove) {
		go UserUnBanned(s, unban, database)
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers // Доп. дозволи

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	logger.Info("The bot is online!")

	sc := make(chan os.Signal, 1) // Вимкнення бота CTRL+C
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
