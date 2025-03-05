package commands

import (
	"context"
	"fmt"
	models2 "telegram-sticker-bot/internal/models"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (c *Commander) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil && (update.Message.Text == "а" || update.Message.Text == "А") {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Хуй на!",
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
	}
	if update.Message != nil && (update.Message.Text == "a" || update.Message.Text == "A") {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Hui na!",
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
	}

	if update.Message != nil && update.Message.ForwardOrigin == nil {
		message := models2.Message{
			ID:            int64(update.Message.ID),
			ChatID:        int64(update.Message.Chat.ID),
			UserID:        update.Message.From.ID,
			SentAt:        time.Now(),
			ReactionCount: 0,
			UserNickname:  update.Message.From.Username,
		}
		c.mdb.Save(message)
	}

	if update.MessageReaction != nil {
		reaction := update.MessageReaction
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

		if bananacount-oldbananacount != 0 {
			err := c.mdb.UpdateBananaReactionCount(reaction.MessageID, reaction.Chat.ID, bananacount-oldbananacount)
			if err != nil {
				fmt.Println(err)
			}
		}
		if likecount-oldlikecount != 0 {
			err := c.mdb.UpdateLikeReactionCount(reaction.MessageID, reaction.Chat.ID, likecount-oldlikecount)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	if update.Message != nil && update.Message.ForwardOrigin != nil {
		userId := update.Message.From.ID
		time.Sleep(time.Second * 2)
		if messagesRaw, ok := UsersForMultiSticker.Load(userId); ok {
			MultiStickerSemaphore.Lock()
			messages := messagesRaw.([]MultiStickerData)
			messages = append(messages, MultiStickerData{
				Message:      update.Message.Text,
				UserNickname: update.Message.ForwardOrigin.MessageOriginUser.SenderUser.Username,
				MessageId:    update.Message.ID,
			})
			UsersForMultiSticker.Store(userId, messages)
			MultiStickerSemaphore.Unlock()
		}
	}
}
