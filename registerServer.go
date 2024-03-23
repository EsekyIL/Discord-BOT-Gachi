package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
	"gopkg.in/natefinch/lumberjack.v2"
)

func registerServer(s *discordgo.Session, g *discordgo.GuildCreate) { // Модуль створення папки серверу, конфігураційного файлу а також лога повідомлень
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
	section.Key("GUILD_REGION").SetValue(g.Guild.Region)
	section = cfg.Section("LOGS")
	section.Key("CHANNEL_LOGS_MESSAGE_ID").SetValue("")
	section.Key("CHANNEL_LOGS_VOICE_ID").SetValue("")
	section.Key("CHANNEL_LOGS_SERVER_ID").SetValue("")
	section = cfg.Section("EMOJI_REACTIONS")
	section.Key("MESSAGE_REACTION_ID").SetValue("")
	section.Key("EMOJI_REACTION").SetValue("")
	section.Key("ROLE_ADD_ID").SetValue("")
	section = cfg.Section("LVL_EXP_USERS")
	members, err := s.GuildMembers(g.Guild.ID, "", 1000)
	if err != nil {
		fmt.Println("Помилка отримання учасників сервера:", err)
		return
	}
	for _, member := range members {
		switch {
		case member.User.ID == "1160175895475138611":
			continue
		case len(member.Roles) == 0:
			continue
		}
		section.Key(member.User.ID).SetValue("0")
	}
	err = cfg.SaveTo(filePath)
	if err != nil {
		errorMsg := fmt.Sprintf("Помилка при збереженні у файл: %v", err)
		writer := color.New(color.FgRed, color.Bold).SprintFunc()
		fmt.Println(writer(errorMsg))
		return
	}
	var logger *log.Logger
	l := &lumberjack.Logger{
		Filename:   "servers/" + g.Guild.ID + "/message.log",
		MaxSize:    8192, // мегабайти
		MaxBackups: 1,
		MaxAge:     30, // дні
	}
	logger = log.New(l, "", log.LstdFlags)
	logger.Println("Привіт, цей бот був написаний ручками 𝕙𝕥𝕥𝕡𝕤://𝕥.𝕞𝕖/𝔼𝕤𝕖𝕜𝕪𝕚𝕝 ♥")
}
