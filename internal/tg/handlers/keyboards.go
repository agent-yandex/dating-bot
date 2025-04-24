package handlers

import "github.com/PaulSonOfLars/gotgbot/v2"

func GetMainReplyKeyboard() gotgbot.ReplyKeyboardMarkup {
	return gotgbot.ReplyKeyboardMarkup{
		Keyboard: [][]gotgbot.KeyboardButton{
			{{Text: "Посмотреть профиль"}},
			{{Text: "Создать/редактировать профиль"}},
			{{Text: "Посмотреть настройки поиска"}},
			{{Text: "Изменить настройки поиска"}},
			{{Text: "Поиск анкет"}},
		},
		ResizeKeyboard: true,
	}
}
