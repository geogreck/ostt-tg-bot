package commands

import (
	"context"
	"fmt"
	"telegram-sticker-bot/internal/auditor"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// for specialized handlers to have priority
	time.Sleep(time.Second * 2)
	if update.Message != nil && update.Message.Video != nil {
		AuditVideoChronHandler(ctx, b, update)
	}
	if update.MessageReaction != nil {
		reaction := update.MessageReaction
		key := fmt.Sprintf("%v/%v", reaction.Chat.ID, reaction.MessageID)

		result := auditor.LoadAuditScore(key)

		likecount, bananacount := 0, 0
		oldlikecount, oldbananacount := 0, 0
		for _, reaction := range update.MessageReaction.NewReaction {
			if reaction.ReactionTypeEmoji.Emoji == "👍" {
				likecount++
			}
			if reaction.ReactionTypeEmoji.Emoji == "🍌" {
				bananacount++
			}
		}
		for _, reaction := range update.MessageReaction.OldReaction {
			if reaction.ReactionTypeEmoji.Emoji == "👍" {
				oldlikecount++
			}
			if reaction.ReactionTypeEmoji.Emoji == "🍌" {
				oldbananacount++
			}
		}
		auditor.StoreAuditScore(key, auditor.AuditScope{BananaCount: result.BananaCount +
			(bananacount - oldbananacount), LikeCount: result.LikeCount + (likecount - oldlikecount)})

		if msgId := auditor.LoadAuditReport(key); msgId != 0 {
			// consider audit already published
			b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    reaction.Chat.ID,
				MessageID: msgId,
				Text:      auditor.BakeAuditReport(auditor.LoadAuditScore(key)),
			})
		}
	}
	userId := update.Message.From.ID
	if messagesRaw, ok := UsersForMultiSticker.Load(userId); ok {
		MultiStickerSemaphore.Lock()
		messages := messagesRaw.([]MultiStickerData)
		messages = append(messages, MultiStickerData{
			update.Message.Text,
			update.Message.ID,
		})
		UsersForMultiSticker.Store(userId, messages)
		MultiStickerSemaphore.Unlock()
	}
}
