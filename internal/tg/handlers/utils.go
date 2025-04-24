package handlers

import (
	"fmt"

	"github.com/agent-yandex/dating-bot/internal/db"
)

// FormatProfile formats a user profile for display.
func FormatProfile(user *db.User, cityName string) string {
	gender := "Не указан"
	if user.Gender == "m" {
		gender = "Мужской"
	} else if user.Gender == "f" {
		gender = "Женский"
	}

	username := "Не указано"
	if user.Username != nil {
		username = *user.Username
	}

	bio := "Не указано"
	if user.Bio != nil {
		bio = *user.Bio
	}

	return fmt.Sprintf(
		"👤 %s\n"+
			"Пол: %s\n"+
			"Возраст: %d\n"+
			"Город: %s\n"+
			"О себе: %s",
		username,
		gender,
		user.Age,
		cityName, // City name requires a join with cities table, omitted for simplicity
		bio,
	)
}

func FormatUserPref(preference *db.UserPreference) string {
	gender := "Не указан"
	if preference.GenderPref == "m" {
		gender = "Мужской"
	} else if preference.GenderPref == "f" {
		gender = "Женский"
	}
	return fmt.Sprintf(
		"Настройки поиска 🔎:\n"+
			"Минимальный возраст: %d\n"+
			"Максимальный возраст: %d\n"+
			"Пол 👤: %s\n"+
			"Область поиска 🌍: %d км",
		preference.MinAge,
		preference.MaxAge,
		gender,
		preference.MaxDistance,
	)
}
