package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"weather-api/internal/adapters/telegram"
	"weather-api/internal/usecase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramController struct {
	bot     *telegram.Bot
	usecase *usecase.WeatherUseCase
}

func NewTelegramController(bot *telegram.Bot, usecase *usecase.WeatherUseCase) *TelegramController {
	return &TelegramController{
		bot:     bot,
		usecase: usecase,
	}
}

func (c *TelegramController) Start(ctx context.Context) error {
	for update := range c.bot.Updates() {
		select {
		case <-ctx.Done():
			return nil
		default:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				c.sendMainMenu(update.Message.Chat.ID)
				continue
			}

			city := update.Message.Text
			if city == "Главное меню" {
				c.sendMainMenu(update.Message.Chat.ID)
				continue
			}

			weather, err := c.usecase.GetWeatherByCity(ctx, city)
			if err != nil {
				c.bot.SendMessage(update.Message.Chat.ID, "Ошибка: город не найден или проблемы с погодой.", nil)
				c.sendMainMenu(update.Message.Chat.ID)
				continue
			}

			weatherResponse := fmt.Sprintf(
				"Погода в %s:\nТемпература: %.1f°C\nСостояние: %s",
				city, weather.CurrentWeather.Temperature, weather.CurrentWeather.WeatherDesc,
			)
			c.bot.SendMessage(update.Message.Chat.ID, weatherResponse, nil)
			c.sendMainMenu(update.Message.Chat.ID)
		}
	}
	return nil
}

func (c *TelegramController) sendMainMenu(chatID int64) {
	cities, err := c.usecase.GetAllCities(context.Background())
	if err != nil {
		slog.Error("failed to get cities", "error", err)
		c.bot.SendMessage(chatID, "Ошибка при загрузке списка городов.", nil)
		return
	}

	var keyboardRows [][]tgbotapi.KeyboardButton
	for i := 0; i < len(cities); i += 2 {
		row := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(cities[i].Name)}
		if i+1 < len(cities) {
			row = append(row, tgbotapi.NewKeyboardButton(cities[i+1].Name))
		}
		keyboardRows = append(keyboardRows, row)
	}
	keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("Главное меню")})
	keyboard := tgbotapi.NewReplyKeyboard(keyboardRows...)

	c.bot.SendMessage(chatID, "Выберите город:", keyboard)
}
