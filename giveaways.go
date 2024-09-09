package main

import (
	"context"
	"encoding/json"
	"fmt"

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

// Функція для збільшення кількості учасників
func incrementParticipantCount(GuildID string, UserID string) (*Giveaway, bool, error) {
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
	giveaway.CountParticipate++

	// Оновлюємо розіграш в Redis
	err = updateGiveaway(giveaway)
	if err != nil {
		Error("error updating giveaway", err)
	}

	return giveaway, false, nil
}
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

func GiveawayCreated(GuildID string, prize string, description string, TimeUnix int64, CountParticipate int, currentTime int, winers string) {
	giveaway := &Giveaway{
		GuildID:          GuildID,
		TimeUnix:         TimeUnix,
		CountParticipate: CountParticipate,
		CurrentTime:      currentTime,
		Description:      description,
		Title:            prize,
		Winers:           winers,
	}

	err := cacheGiveaway(giveaway)
	if err != nil {
		fmt.Println("Error caching giveaway:", err)
	}
}
