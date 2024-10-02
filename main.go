package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var usedCodes = make(map[string]bool) // Зберігаємо всі згенеровані коди

type RowData struct {
	guild_id           string
	guild_name         string
	guild_owner_id     string
	guild_owner_name   string
	members            string
	vip                bool
	forum              bool
	channel_id_forum   string
	channel_id_message string
	channel_id_voice   string
	channel_id_server  string
	channel_id_penalty string
}

var start_time float64
var (
	cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_usage",
		Help: "CPU usage percentage.",
	})
	memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "memory_usage",
		Help: "Memory usage in MB.",
	})
	startTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "start_time",
		Help: "Unix timestamp of when the application started.",
	})
)

func init() {
	// Реєструємо метрики
	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(memoryUsage)
	prometheus.MustRegister(startTime)
	start_time = float64(time.Now().Unix())
	startTime.Set(float64(time.Now().Unix()))
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

func SelectDB(query string, args ...interface{}) (*RowData, error) {
	var result RowData
	database, err := ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}
	defer database.Close()

	columnName := "guild_id" // Замініть на назву стовпця, який потрібно перевірити

	exists, err := ColumnExists(database, columnName)
	if err != nil {
		return nil, err
	}

	if !exists {
		err := fmt.Errorf("no rows found")
		return nil, err
	}

	// Використовуємо підготовлений запит із параметрами
	statement, err := database.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %v, query: %s", err, query)
	}
	defer statement.Close()

	// Виконуємо запит з переданими параметрами
	rows, err := statement.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v, query: %s", err, query)
	}
	defer rows.Close()

	// Проходимо через результати
	if rows.Next() {
		err := rows.Scan(
			&result.guild_id,
			&result.guild_name,
			&result.guild_owner_id,
			&result.guild_owner_name,
			&result.members,
			&result.vip,
			&result.forum,
			&result.channel_id_forum,
			&result.channel_id_message,
			&result.channel_id_voice,
			&result.channel_id_server,
			&result.channel_id_penalty,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
	} else {
		return nil, fmt.Errorf("no rows found")
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}
func UpdateDB(query string, args ...interface{}) error {

	database, err := ConnectDB()
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}
	defer database.Close()

	// Використовуємо підготовлений запит із параметрами
	statement, err := database.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v, query: %s", err, query)
	}
	defer statement.Close()

	// Виконуємо запит з переданими параметрами
	_, err = statement.Exec(args...)
	if err != nil {
		return fmt.Errorf("error executing query: %v, query: %s", err, query)
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
func corsMiddleware(next http.Handler) http.Handler { // The function of permissions for monitoring, if it is not needed - delete it.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")             // Дозволяє всі походження
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // Дозволяє методи
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Дозволяє заголовки
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	go func() {
		http.Handle("/metrics", corsMiddleware(promhttp.Handler()))
		fmt.Println(time.Now().Format(time.RFC1123), "Prometheus metrics available at :8081/metrics")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()
	go func() {
		fs := http.FileServer(http.Dir("./localhost"))
		http.Handle("/", fs)
		fmt.Println(time.Now().Format(time.RFC1123), "Monitoring web-site avaliable at :3000/")
		log.Fatal(http.ListenAndServe(":3000", nil))
	}()

	// Two functions that raise local servers for monitoring from above. You can remove these features.

	token := goDotEnvVariable("API_KEY")
	sess, _ := discordgo.New("Bot " + token)

	registerCommands(sess)

	sess.StateEnabled = true
	sess.State.MaxMessageCount = 1000
	sess.State.TrackVoice = true
	sess.State.TrackMembers = true
	sess.State.TrackRoles = true

	sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		go registerServer(s, g) // Виклик функції для реєстрації сервера, якщо дані не знайдено

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

	sess.AddHandler(func(s *discordgo.Session, cc *discordgo.ChannelCreate) {
		go ChannelCreated(s, cc)
	})

	sess.AddHandler(func(s *discordgo.Session, cd *discordgo.ChannelDelete) {
		go ChannelDeleted(s, cd)
	})

	go func() {
		// Запускаємо перевірку
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			checkGiveaways(sess)
		}
	}()

	go func() { // The monitoring function can be removed.
		for {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			memoryUsage.Set(float64(memStats.Alloc) / 1024 / 1024)

			cpuPercentages, err := cpu.Percent(0, false)
			if err != nil {
				log.Printf("Error getting CPU usage: %v", err)
			} else {
				if len(cpuPercentages) > 0 {
					cpuUsage.Set(cpuPercentages[0])
				}
			}

			startTime.Set(start_time)

			time.Sleep(1 * time.Second)
		}
	}()

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers // Доп. дозволи

	err := sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println(time.Now().Format(time.RFC1123), "The bot is online!")

	sc := make(chan os.Signal, 1) // Вимкнення бота CTRL+C
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
