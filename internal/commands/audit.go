package commands

import (
	"context"
	"fmt"
	"strings"
	"telegram-sticker-bot/internal/auditor"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func AuditVideoHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Video == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ответьте на сообщение с видео, чтобы пройти аудит.",
		})
		return
	}
	// update.MessageReaction.
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Аудит пройден",
	})
}

func AuditVideoChronHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Video == nil {
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: "Данное сообщение было распознано, как потенциально очень смешной тикток" +
			" и было отправлено на прохождение аудита.",
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})

	go func() {
		time.Sleep(time.Second * 10)

		key := fmt.Sprintf("%v/%v", update.Message.Chat.ID, update.Message.ID)
		result := auditor.LoadAuditScore(key)

		msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   auditor.BakeAuditReport(result),
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		if err != nil {
			fmt.Println("Не получилось отправить сообщение с аудитом: ", err)
			return
		}
		auditor.StoreAuditReport(key, msg.ID)
	}()
}

func AuditTopVideosHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	top, err := auditor.GetTopAuditKeys(10)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при получении топа: " + err.Error(),
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	msg := ""
	for i, pos := range top {
		data := strings.Split(pos, "/")
		link := fmt.Sprintf("%d: https://t.me/c/%v/%v", i+1, data[0][4:], data[1])
		msg += link + "\n"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})
}
