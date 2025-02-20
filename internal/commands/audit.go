package commands

import (
	"context"
	"fmt"
	"telegram-sticker-bot/internal/auditor"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (c *Commander) AuditHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.ReplyToMessage == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ответьте на сообщение, чтобы провести аудит.",
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}

	bananaCount, likeCount, err := c.mdb.GetReactions(update.Message.ReplyToMessage.ID, update.Message.Chat.ID)
	if err != nil {
		fmt.Printf("Error fetching reactions: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не получилось подготовить аудитный отчет для этого сообщения: " + err.Error(),
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: auditor.BakeAuditReport(auditor.AuditScope{
			BananaCount: bananaCount,
			LikeCount:   likeCount,
		}),
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})
	if err != nil {
		fmt.Printf("Error sending reactions report: %v", err)
	}
}

func AuditTopVideosHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

}
