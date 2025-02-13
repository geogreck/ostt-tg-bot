package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"telegram-sticker-bot/internal/commands"

	"github.com/go-telegram/bot"
)

// Имя стикер пака должно быть уникальным и соответствовать требованиям Telegram.
// Например: имя_пакета_by_<botusername>
const (
	StickerSetName  = "ostt_by_OsttBotAuditor" // замените mybot на имя вашего бота
	StickerSetTitle = "Пак очень смешных тиктоков"
)

func main() {
	// Получаем токен из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Необходимо установить переменную окружения TELEGRAM_BOT_TOKEN")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithAllowedUpdates(bot.AllowedUpdates{"message", "message_reaction"}),
		bot.WithDefaultHandler(commands.DefaultHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/sticker", bot.MatchTypePrefix, commands.AddStickerHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/poll", bot.MatchTypePrefix, commands.PollHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/createset", bot.MatchTypePrefix, commands.CreateStickerSetHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/audit", bot.MatchTypePrefix, commands.AuditVideoHandler)

	b.Start(ctx)
}
