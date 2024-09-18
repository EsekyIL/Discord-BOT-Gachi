package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client
var ctx = context.Background()

type Giveaway struct {
	GuildID          string   `json:"guild_id"`
	Description      string   `json:"description"`
	TimeUnix         int64    `json:"time_unix"`
	CountParticipate int      `json:"count_participate"`
	CurrentTime      int      `json:"current_time"`
	UserParticipate  []string `json:"user_participate"`
	Title            string   `json:"title"`
	Winers           string   `json:"winers"`
	ChannelID        string   `json:"channel_id"`
	MessageID        string   `json:"message_id"`
	AuthorID         string   `json:"author_id"`
}

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // адреса Redis-сервера
	})

	// Перевірка з'єднання
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to Redis: %v", err))
	}
}

func cacheGiveaway(giveaway *Giveaway) error {
	giveawayJSON, err := json.Marshal(giveaway)
	if err != nil {
		return fmt.Errorf("failed to marshal giveaway: %v", err)
	}

	key := fmt.Sprintf("giveaway:%s", giveaway.GuildID)
	err = rdb.Set(ctx, key, giveawayJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set giveaway in cache: %v", err)
	}

	return nil
}

// Функція для додавання розіграшу в відсортовану множину за часом закінчення
func addGiveawayToSortedSet(giveaway *Giveaway) error {
	key := "giveaways:endtimes"
	err := rdb.ZAdd(ctx, key, &redis.Z{
		Score:  float64(giveaway.TimeUnix), // Час закінчення як score
		Member: giveaway.GuildID,           // Ідентифікатор розіграшу як елемент
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add giveaway to sorted set: %v", err)
	}
	return nil
}

// Функція для оновлення розіграшу в Redis
func updateGiveaway(giveaway *Giveaway) error {
	giveawayJSON, err := json.Marshal(giveaway)
	if err != nil {
		return fmt.Errorf("failed to marshal giveaway: %v", err)
	}

	key := fmt.Sprintf("giveaway:%s", giveaway.GuildID)
	err = rdb.Set(ctx, key, giveawayJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set giveaway in cache: %v", err)
	}

	return nil
}

// Функція для вибірки розіграшів, що скоро закінчаться
func getGiveawaysEndingSoon(currentTime int64) ([]string, error) {
	key := "giveaways:endtimes"
	guildIDs, err := rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "0",                            // Початок діапазону (початок часів)
		Max: fmt.Sprintf("%d", currentTime), // Кінець діапазону (поточний час)
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get giveaways ending soon: %v", err)
	}
	return guildIDs, nil
}

// Видалення розіграшу з відсортованої множини
func removeFinishedGiveaway(guildID string) error {
	key := "giveaways:endtimes"
	err := rdb.ZRem(ctx, key, guildID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove finished giveaway: %v", err)
	}
	return nil
}
func leaveUserGiveaway(GuildID string, UserID string) (bool, error) {
	giveaway, err := getGiveaway(GuildID)
	if err != nil {
		Error("error getting giveaway", err)
	}

	if giveaway.UserParticipate == nil {
		giveaway.UserParticipate = []string{}
	}

	for i, user := range giveaway.UserParticipate {
		if user == UserID {
			giveaway.UserParticipate = append(giveaway.UserParticipate[:i], giveaway.UserParticipate[i+1:]...)
			err = updateGiveaway(giveaway)
			if err != nil {
				Error("error updating giveaway", err)
			}
			return true, nil
		}
	}

	return false, nil
}

// Функція для збільшення кількості учасників
func incrementParticipantCount(GuildID string, UserID string, MessageID string) (*Giveaway, bool, error) {
	// Отримуємо існуючий розіграш
	giveaway, err := getGiveaway(GuildID)
	if err != nil {
		Error("error getting giveaway", err)
	}
	// Ініціалізуємо зріз UserParticipate, якщо він ще не ініціалізований
	if giveaway.UserParticipate == nil {
		giveaway.UserParticipate = []string{}
	}
	// Перевіряємо, чи вже є UserID у зрізі UserParticipate
	for _, user := range giveaway.UserParticipate {
		if user == UserID {
			// Якщо користувач вже є, повертаємо помилку або нічого не робимо
			return giveaway, true, nil
		}
	}

	// Додаємо нового учасника
	giveaway.UserParticipate = append(giveaway.UserParticipate, UserID)
	giveaway.MessageID = MessageID
	giveaway.CountParticipate++

	// Оновлюємо розіграш в Redis
	err = updateGiveaway(giveaway)
	if err != nil {
		Error("error updating giveaway", err)
	}

	return giveaway, false, nil
}

// Отримання розіграшу
func getGiveaway(GuildID string) (*Giveaway, error) {
	key := fmt.Sprintf("giveaway:%s", GuildID)
	giveawayJSON, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("giveaway not found for GuildID: %s", GuildID)
		}
		return nil, fmt.Errorf("failed to get giveaway from cache: %v", err)
	}

	var giveaway Giveaway
	err = json.Unmarshal([]byte(giveawayJSON), &giveaway)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal giveaway: %v", err)
	}

	return &giveaway, nil
}

// Функція для створення нового розіграшу
func GiveawayCreated(GuildID string, prize string, description string, TimeUnix int64, CountParticipate int, currentTime int, winers string, channelID string, author string) {
	giveaway := &Giveaway{
		GuildID:          GuildID,
		TimeUnix:         TimeUnix,
		CountParticipate: CountParticipate,
		CurrentTime:      currentTime,
		Description:      description,
		Title:            prize,
		Winers:           winers,
		ChannelID:        channelID,
		AuthorID:         author,
	}

	// Кешуємо розіграш
	err := cacheGiveaway(giveaway)
	if err != nil {
		fmt.Println("Error caching giveaway:", err)
	}

	// Додаємо розіграш до відсортованої множини
	err = addGiveawayToSortedSet(giveaway)
	if err != nil {
		fmt.Println("Error adding giveaway to sorted set:", err)
	}
}

func randomInts(min, max, count int) []int {
	if min >= max {
		return nil
	}
	results := make([]int, count)
	for i := 0; i < count; i++ {
		results[i] = rand.Intn(max-min+1) + min
	}
	return results
}
func GiveawayEnd(giveaway *Giveaway, s *discordgo.Session) {
	min := 0
	max := len(giveaway.UserParticipate) - 1
	count, err := strconv.Atoi(giveaway.Winers)
	if err != nil {
		Error("error end giveavay", err)
		return
	}

	winners := randomInts(min, max, count)
	if winners == nil {
		return
	}

	var rsp strings.Builder

	for i, winner := range winners {
		if winner > len(giveaway.UserParticipate) {
			break
		}
		if i == count-1 {
			temp := "<@" + giveaway.UserParticipate[winner] + ">. "
			rsp.WriteString(temp)
			break
		} else {
			temp := "<@" + giveaway.UserParticipate[winner] + ">, "
			rsp.WriteString(temp)
		}
	}
	embed := &discordgo.MessageEmbed{
		Title: "Contest winners",
		Color: 0xfadb84,
		Description: fmt.Sprintf("Congratulations %s \n"+"You won the "+"`%s.`",
			rsp.String(), giveaway.Title,
		),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, err = s.ChannelMessageSendEmbedReply(giveaway.ChannelID, embed, &discordgo.MessageReference{
		MessageID: giveaway.MessageID,
		GuildID:   giveaway.GuildID,
		ChannelID: giveaway.ChannelID,
	})
	if err != nil {
		Error("error message update", err)
		return
	}
	embed = &discordgo.MessageEmbed{
		Title: giveaway.Title,
		Color: 0xfadb84,
		Description: fmt.Sprintf(giveaway.Description+"\n\n"+">>> **Ended: **"+"<t:%d:R>"+"  "+"<t:%d:f>"+"\n"+"** Hosted by: **"+"<@%s>"+"\n"+"**Entries: **"+"`%d`"+"\n"+"**Winner: **"+"%s",
			giveaway.TimeUnix, giveaway.TimeUnix, giveaway.AuthorID, giveaway.CountParticipate, rsp.String(),
		),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.PrimaryButton,
					Disabled: true,
					CustomID: "participate",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎆",
					},
				},
			},
		},
	}
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    giveaway.ChannelID,
		ID:         giveaway.MessageID,
		Embed:      embed,
		Components: &components, // Тут передаємо слайс без вказівника
	})
	if err != nil {
		Error("Channel message edit complex", err)
		return
	}

}

// Перевірка розіграшів і обробка закінчення
func checkGiveaways(s *discordgo.Session) {
	currentTime := time.Now().Unix()
	endingGiveaways, err := getGiveawaysEndingSoon(currentTime)
	if err != nil {
		Error("Error fetching giveaways", err)
		return
	}
	// Якщо немає розіграшів, повертаємося
	if len(endingGiveaways) == 0 {
		return
	}

	for _, guildID := range endingGiveaways {
		// Отримуємо деталі розіграшу
		giveaway, err := getGiveaway(guildID)
		if err != nil {
			Error("Error getting giveaway", err)
			continue
		}

		// Обробляємо закінчення розіграшу
		GiveawayEnd(giveaway, s)

		// Видаляємо розіграш з відсортованої множини
		err = removeFinishedGiveaway(guildID)
		if err != nil {
			Error("Error removing finished giveaway", err)
		}
	}
}
