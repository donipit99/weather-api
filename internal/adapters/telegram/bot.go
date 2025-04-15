package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"weather-api/internal/dto"
	"weather-api/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WeatherUseCase interface {
	GetWeatherByCity(ctx context.Context, cityName string) (*dto.WeatherResult, error)
	GetAllCities(ctx context.Context) ([]models.City, error)
}

type Bot struct {
	botAPI  *tgbotapi.BotAPI
	useCase WeatherUseCase
	updates tgbotapi.UpdatesChannel
}

func NewBot(token string, useCase WeatherUseCase) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	botAPI.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	return &Bot{
		botAPI:  botAPI,
		useCase: useCase,
		updates: updates,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	for update := range b.updates {
		select {
		case <-ctx.Done():
			return nil
		default:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				b.sendMainMenu(update.Message.Chat.ID)
				continue
			}

			city := update.Message.Text
			if city == "Главное меню" {
				b.sendMainMenu(update.Message.Chat.ID)
				continue
			}

			weather, err := b.useCase.GetWeatherByCity(ctx, city)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: город не найден или проблемы с погодой.")
				b.botAPI.Send(msg)
				b.sendMainMenu(update.Message.Chat.ID)
				continue
			}

			response := fmt.Sprintf(
				"Погода в %s:\nТемпература: %.1f°C\nСостояние: %s",
				city, weather.CurrentWeather.Temperature, weather.CurrentWeather.WeatherDesc,
			)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			b.botAPI.Send(msg)
			b.sendMainMenu(update.Message.Chat.ID)
		}
	}
	return nil
}

func (b *Bot) sendMainMenu(chatID int64) {
	cities, err := b.useCase.GetAllCities(context.Background())
	if err != nil {
		slog.Error("failed to get cities", "err", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при загрузке списка городов.")
		b.botAPI.Send(msg)
		return
	}

	// Создаем клавиатуру
	var keyboardRows [][]tgbotapi.KeyboardButton
	for i := 0; i < len(cities); i += 2 {
		row := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(cities[i].Name)}
		if i+1 < len(cities) {
			row = append(row, tgbotapi.NewKeyboardButton(cities[i+1].Name))
		}
		keyboardRows = append(keyboardRows, row)
	}
	// Добавляем кнопку "Главное меню"
	keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("Главное меню")})

	keyboard := tgbotapi.NewReplyKeyboard(keyboardRows...)

	msg := tgbotapi.NewMessage(chatID, "Выберите город:")
	msg.ReplyMarkup = keyboard
	b.botAPI.Send(msg)
}

func (b *Bot) Stop() {
	b.botAPI.StopReceivingUpdates()
}
