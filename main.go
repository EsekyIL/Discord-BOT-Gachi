package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

// Колір помилок commands.go - червоний
func goDotEnvVariable(key string) string {

	// завантажити файл .env
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Помилка завантаження файлу .env")
	}
	return os.Getenv(key)
}

func main() {
	token := goDotEnvVariable("API_KEY")
	userChannels := make(map[string]string)
	userTimeJoinVoice := make(map[string]string)
	sess, err := discordgo.New("Bot " + token) // Відкриття сессії з ботом
	if err != nil {
		log.Fatal(err)
	}
	registerCommands(sess)
	sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		basePath := "./servers"
		folderName := g.Guild.ID
		folderPath := filepath.Join(basePath, folderName)
		_, err := os.Stat(folderPath)
		if os.IsNotExist(err) {
			registerServer(g)
			red := color.New(color.FgRed)
			boldRed := red.Add(color.Bold)
			whiteBackground := boldRed.Add(color.BgCyan)
			whiteBackground.Printf("🎉 Урааа. %v добавили бота на свій сервер! 🎉", g.Guild.Name)
			print("\n")
		}
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) { // Модуль відстеження повідомлень, а також запис їх у log!
		if m.Author.Bot {
			return
		}
		MessageSaveToLog(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) { // Модуль додавання ролі по реакції на повідомлення
		//RoleAddByEmoji(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) { // Модуль логування оновленого повідомлення, а також запис у log
		if m.Author == nil || m.Author.Bot {
			return
		}
		MessageUpdateToLog(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) { // Модуль логування видаленого повідомлення
		MessageDeleteLog(s, m)
	})
	sess.AddHandler(func(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) { // Модуль логування входу/переходу/виходу в голосових каналах
		VoiceLog(s, vs, &userChannels, &userTimeJoinVoice)
	})
	sess.AddHandler(func(s *discordgo.Session, gma *discordgo.GuildMemberAdd) { // Модуль логування надходження користувачів на сервер
		InvitePeopleToServer(s, gma)
	})
	sess.AddHandler(func(s *discordgo.Session, gmr *discordgo.GuildMemberRemove) { // Модуль логування виходу користувачів з серверу
		ExitPeopleFromServer(s, gmr)
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
		BanUserToServer(s, b, ba)
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
