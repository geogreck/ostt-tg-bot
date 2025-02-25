package commands

import (
	"context"
	"fmt"
	"strings"
	"telegram-sticker-bot/internal/auditor"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/goodsign/monday"
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

func (c *Commander) AuditTopHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	messages, err := c.mdb.GetTopMessages(update.Message.Chat.ID)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не получилось подготовить бухгалтерский отчет по наиболее смешным сообщениям в чате: " + err.Error(),
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}
	report := "Лучшие сообщения в чате за всё время:\n\n"
	chatId := -1*update.Message.Chat.ID - 1000000000000

	for id, message := range messages {
		report += fmt.Sprintf("%d\\. [@%v %v](https://t.me/c/%v/%v)\n", id+1,
			strings.Replace(message.UserNickname, "_", "\\_", -1), monday.Format(message.SentAt, "2 January", monday.LocaleRuRU), chatId, message.ID,
		)
	}
	fmt.Println()
	fmt.Println(report)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   report,
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		fmt.Printf("Error top messages report: %v", err)
	}
}
