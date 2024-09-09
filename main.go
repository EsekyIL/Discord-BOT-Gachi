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

type Translations struct {
	MessageDeleted         string
	MessageUpdated         string
	MessageContent         string
	Was                    string
	NowIs                  string
	Channel                string
	Author                 string
	MessageID              string
	UserJoinVoice          string
	UserLeftVoice          string
	UserMovedVoice         string
	User                   string
	OldRoom                string
	NewRoom                string
	InviteCreated          string
	Code                   string
	ValidityPeriod         string
	CountUser              string
	NewUser                string
	Created                string
	ID                     string
	Kick                   string
	Reason                 string
	UserLeftGuild          string
	UnMute                 string
	Mute                   string
	TimeRemove             string
	Ban                    string
	UnBan                  string
	SettingFunction        string
	SelectItem             string
	SelectSettingItem      string
	Logging                string
	EventLogging           string
	Lang                   string
	ChangeLang             string
	ConfigLogging          string
	ChannelsLog            string
	SelectFirstChannel     string
	FirstChannelDescrip    string
	SelectSecondChannel    string
	SecondChannelDescrip   string
	SelectThirdChannel     string
	ThirdChannelDescrip    string
	SelectChannel          string
	ChannelDescrip         string
	Placeholder            string
	IfChangeLang           string
	BigDescrip             string
	AllLogs                string
	Message                string
	VoiceChannels          string
	Events                 string
	Success                string
	UseAllLogs             string
	UseAllLogsFirstChannel string
	UseMessageLog          string
	UseVoiceLog            string
}

var translations = map[string]Translations{
	"EU": {
		MessageDeleted:         "Deleted message",
		MessageUpdated:         "Message updated",
		MessageContent:         "Content",
		Was:                    "Was",
		NowIs:                  "Now is",
		Channel:                "Channel",
		Author:                 "Author",
		MessageID:              "Message ID",
		UserJoinVoice:          "The user entered the room",
		UserLeftVoice:          "The user left the room",
		UserMovedVoice:         "The user moved to the second room",
		User:                   "User",
		OldRoom:                "Old room",
		NewRoom:                "New room",
		InviteCreated:          "An invitation has been created",
		Code:                   "Code",
		ValidityPeriod:         "Validity period",
		CountUser:              "Count of users",
		NewUser:                "New user",
		Created:                "Created",
		ID:                     "ID",
		Kick:                   "Kick",
		Reason:                 "Reason",
		UserLeftGuild:          "The user has left the guild",
		UnMute:                 "Unmuted",
		Mute:                   "Muted",
		TimeRemove:             "Time to remove the restriction",
		Ban:                    "Ban",
		SettingFunction:        "Setting up bot functions",
		SelectItem:             "Select the item you want to adjust",
		SelectSettingItem:      "Select the setting item 👇",
		Logging:                "Logging in",
		EventLogging:           "Event logging on the server",
		Lang:                   "Language",
		ChangeLang:             "Language change",
		ConfigLogging:          "Configuring server logging",
		ChannelsLog:            "Select channels to log. From `one' to `three'. You can change them at any time.",
		SelectFirstChannel:     "Select the first channel",
		FirstChannelDescrip:    "*The first channel shows you `change/delete` messages on the server.*",
		SelectSecondChannel:    "Select the second channel",
		SecondChannelDescrip:   "*The second channel gives you `enter/transition/exit` from the voice channels on the server.*",
		SelectThirdChannel:     "Selecting the third channel",
		ThirdChannelDescrip:    "*The third channel displays `login/logout/ban` of the user on the server.*",
		SelectChannel:          "Channel selection",
		ChannelDescrip:         "***If you want the logging output to one channel, just select the channel you need!***",
		Placeholder:            "It is necessary to poke here",
		IfChangeLang:           ">>> *If you want to change the language, press the button!*",
		BigDescrip:             ">>> *If you want to send all logs to one channel, click the `All logs` button. If you need specific logging for messages, voice channels, or events, select the appropriate option.*",
		AllLogs:                "All logs",
		Message:                "Message",
		VoiceChannels:          "Voice channels",
		Events:                 "Events",
		Success:                "Successfully",
		UseAllLogs:             "> You can now use server-wide logging",
		UseAllLogsFirstChannel: "> Now you can use logging of the whole server in only one channel",
		UseMessageLog:          "> Now you can only use message logging",
		UseVoiceLog:            "> Now you can only use voice channel logging",
	},
	"UA": {
		MessageDeleted:         "Видалено повідомлення",
		MessageUpdated:         "Повідомлення оновлено",
		MessageContent:         "Вміст",
		Was:                    "Було",
		NowIs:                  "Стало",
		Channel:                "Канал",
		Author:                 "Автор",
		MessageID:              "Айді повідомлення",
		UserJoinVoice:          "Користувач зайшов у кімнату",
		UserLeftVoice:          "Користувач вийшов з кімнати",
		UserMovedVoice:         "Користувач перейшов у другу кімнату",
		User:                   "Користувач",
		OldRoom:                "Стара кімната",
		NewRoom:                "Нова кімната",
		InviteCreated:          "Створено запрошення",
		Code:                   "Код",
		ValidityPeriod:         "Термін дії",
		CountUser:              "К-сть користувачів",
		NewUser:                "Новий користувач",
		Created:                "Створений",
		ID:                     "Айді",
		Kick:                   "Кік",
		Reason:                 "Причина",
		UserLeftGuild:          "Користувач покинув гільдію",
		UnMute:                 "Увімкнено чат",
		TimeRemove:             "Час зняття обмеження",
		Mute:                   "Мут",
		Ban:                    "Бан",
		SettingFunction:        "Налаштування функцій бота",
		SelectItem:             "Виберіть пункт, який хочете налаштувати",
		SelectSettingItem:      "Виберіть пункт налаштування 👇",
		Logging:                "Логування",
		EventLogging:           "Логування подій на сервері",
		Lang:                   "Мова",
		ChangeLang:             "Зміна мови",
		ConfigLogging:          "Налаштування логування серверу",
		ChannelsLog:            "Виберіть канали для логування. Від `одного` до `трьох`. У будь який момент їх можна змінити.",
		SelectFirstChannel:     "Вибір першого каналу",
		FirstChannelDescrip:    "*Перший канал виводить вам `зміну/видалення` повідомлень на сервері.*",
		SelectSecondChannel:    "Вибір другого каналу",
		SecondChannelDescrip:   "*Другий канал виводить вам `вхід/перехід/вихід` з голосових каналів на сервері.*",
		SelectThirdChannel:     "Вибір третього каналу",
		ThirdChannelDescrip:    "*Третій канал виводить `вхід/вихід/бан` користувача на сервері.*",
		SelectChannel:          "Вибір каналу",
		ChannelDescrip:         "***Якщо хочете, щоб логування виводилось в один канал, просто виберіть той канал, який вам потрібен!***",
		Placeholder:            "Тута треба тицьнути",
		IfChangeLang:           ">>> *Якщо ви бажаєте змінити мову - натисніть кнопку.*",
		BigDescrip:             ">>> *Якщо ви хочете всі логи направляти до одного каналу, натисніть кнопку `Усі логи`. Якщо вам потрібне конкретне логування для повідомлень, голосових каналів або подій, виберіть відповідну опцію.*",
		AllLogs:                "Усі логи",
		Message:                "Повідомлення",
		VoiceChannels:          "Голосові канали",
		Events:                 "Події",
		Success:                "Успішно",
		UseAllLogs:             "> Тепер можете користуватись логуванням всього серверу.",
		UseAllLogsFirstChannel: "> Тепер можете користуватись логуванням всього серверу лише в один канал.",
		UseMessageLog:          "> Тепер можете користуватись тільки логуванням повідомлень.",
		UseVoiceLog:            "> Тепер можете користуватись тільки логуванням голосових каналів.",
	},
}

func getTranslation(lang string) Translations {
	return translations[lang]
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

func SelectDB(select_row string, GuildID string) (int, string) {
	database, _ := ConnectDB()

	query := fmt.Sprintf("SELECT id, %s, Language FROM %s WHERE id = ?", select_row, shortenNumber(GuildID))
	rows, err := database.Query(query, GuildID)
	if err != nil {
		Error("Щось сталось", err)
		return 0, "" // Повертаємо 0 у разі помилки
	}
	defer rows.Close()

	var id int
	var channel_log int
	var lang string

	if rows.Next() {
		err := rows.Scan(&id, &channel_log, &lang)
		if err != nil {
			Error("Failed to scan the row", err)
			return 0, ""
		}
	} else {
		if err := rows.Err(); err != nil {
			Error("Failed during iteration over rows", err)
		}
		return 0, ""
	}

	return channel_log, lang
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
		go registerServer(g, database) // Виклик функції для реєстрації сервера, якщо дані не знайдено

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
		// Отримання аудиторських записів для сервера
		if mute.BeforeUpdate == nil || mute.CommunicationDisabledUntil == nil {
			return
		}
		go UserMuted(s, mute)
	})
	sess.AddHandler(func(s *discordgo.Session, unban *discordgo.GuildBanRemove) {
		go UserUnBanned(s, unban)
	})

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
