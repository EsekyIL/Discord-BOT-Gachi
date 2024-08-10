package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

func registerServer(g *discordgo.GuildCreate) { // Модуль створення папки серверу, конфігураційного файлу а також лога повідомлень
	basePath := "./servers"
	folderName := g.Guild.ID
	folderPath := filepath.Join(basePath, folderName)
	// Перевірка, чи існує каталог за вказаним шляхом
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		// Якщо каталог не існує, створити його
		err = os.Mkdir(folderPath, 0755)
		if err != nil {
			// Обробка помилки при створенні каталогу
			fmt.Println("Помилка при створенні каталогу:", err)
			return
		}
	} else if err != nil {
		// Обробка іншої можливої помилки при перевірці існування каталогу
		fmt.Println("Помилка при перевірці каталогу:", err)
		return
	}
	directoryPath := filepath.Join(basePath, folderName)
	filePath := filepath.Join(directoryPath, "config.ini")
	cfg := ini.Empty()
	section := cfg.Section("GUILD")
	section.Key("GUILD_NAME").SetValue(g.Guild.Name)
	section.Key("GUILD_ID").SetValue(g.Guild.ID)
	section.Key("GUILD_MEMBERS").SetValue(strconv.Itoa(g.Guild.MemberCount))
	section = cfg.Section("LOGS")
	section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue("")
	section.Key("CHANNEL_LOGS_VOICE_ID").SetValue("")
	section.Key("CHANNEL_LOGS_SERVER_ID").SetValue("")
	section = cfg.Section("EMOJI_REACTIONS")
	section.Key("MESSAGE_REACTION_ID").SetValue("")
	section.Key("EMOJI_REACTION").SetValue("")
	section.Key("ROLE_ADD_ID").SetValue("")
	err = cfg.SaveTo(filePath)
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при збереженні у файл: %v", err)
		writer := color.New(color.FgRed, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
		return
	}

	logFilePath := "servers/" + g.Guild.ID + "/message.json"
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		slog.Error("Не вдалося відкрити файл для логів", "error", err)
		return
	}
	defer file.Close()

}
