package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	botAPI  *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

func NewBot(token string) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	botAPI.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := botAPI.GetUpdatesChan(updateConfig)

	return &Bot{
		botAPI:  botAPI,
		updates: updates,
	}, nil
}

func (b *Bot) Updates() tgbotapi.UpdatesChannel {
	return b.updates
}

func (b *Bot) SendMessage(chatID int64, text string, replyMarkup interface{}) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if replyMarkup != nil {
		msg.ReplyMarkup = replyMarkup
	}
	_, err := b.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (b *Bot) Stop() {
	b.botAPI.StopReceivingUpdates()
}
