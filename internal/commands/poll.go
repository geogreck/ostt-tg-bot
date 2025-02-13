package commands

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func PollHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println(update.Message)
	p := &bot.SendPollParams{
		ChatID:   update.Message.Chat.ID,
		Question: "Сосал?",
		Options: []models.InputPollOption{
			{
				Text: "Да",
			},
			{
				Text: "Нет",
			},
		},
		IsAnonymous: bot.False(),
	}
	b.SendPoll(ctx, p)
}
