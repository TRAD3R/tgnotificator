package tgnotificator

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	rps      = 20
	interval = time.Minute
)

type Telegram struct {
	bot         *tgbotapi.BotAPI
	channel     int64
	logger      *slog.Logger
	serviceName string
	rateLimiter *rateLimiter
}

func NewTelegram(token string, channel int64, logger *slog.Logger, serviceName string, isDebug bool) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to bot connect: %w", err)
	}

	bot.Debug = isDebug

	logger.Debug("start telegram")
	return &Telegram{
		bot:         bot,
		channel:     channel,
		logger:      logger,
		serviceName: serviceName,
		rateLimiter: newRateLimiter(rps, interval),
	}, nil
}

func (t *Telegram) SendMessage(msg string) {
	if !t.rateLimiter.allowRequest() {
		requests, remainingTime := t.rateLimiter.currentState()
		t.logger.Debug("limit request", "requests", requests, "remaining", remainingTime.String())
		time.Sleep(remainingTime)
	}

	msg = fmt.Sprintf("%s:\n%s", t.serviceName, msg)
	message := tgbotapi.NewMessage(t.channel, msg)
	message.ParseMode = tgbotapi.ModeHTML

	if _, err := t.bot.Send(message); err != nil {
		t.logger.Error("failed to send message", "err", err.Error())
	}
}

func (t *Telegram) SendFile(filepath string, msg string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.logger.Error("failed to close file", "file", file.Name(), "err", err.Error())
		}
	}()

	fileReader := tgbotapi.FileReader{
		Name:   file.Name(),
		Reader: file,
	}

	message := tgbotapi.NewDocument(t.channel, fileReader)
	if msg != "" {
		message.Caption = fmt.Sprintf("%s: %s", t.serviceName, msg)
	}
	if _, err := t.bot.Send(message); err != nil {
		t.logger.Error("failed to send file", "err", err.Error())
	}

	return nil
}
