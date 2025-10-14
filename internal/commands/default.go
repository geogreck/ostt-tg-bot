package commands

import (
    "context"
    "fmt"
    "math/rand"
    "net/http"
    "strings"
    models2 "telegram-sticker-bot/internal/models"
    "time"

    "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

var autoAskSent syncMap

type syncMap struct{ m map[string]struct{} }

func (s *syncMap) LoadOrStore(k string) bool {
    if s.m == nil {
        s.m = make(map[string]struct{})
    }
    if _, ok := s.m[k]; ok {
        return true
    }
    s.m[k] = struct{}{}
    return false
}

func init() {
    rand.Seed(time.Now().UnixNano())
}

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

    // Auto-ask: 5% chance for qualifying messages (length>25 or ends with '?'), only in allowed chat, once per message
    if update.Message != nil && update.Message.Text != "" {
        const allowedChatShortID int64 = 2115621645
        chatID := update.Message.Chat.ID
        isAllowed := chatID == allowedChatShortID || (-chatID-1000000000000) == allowedChatShortID
        if isAllowed {
            text := strings.TrimSpace(update.Message.Text)
            longEnough := len([]rune(text)) > 25
            endsWithQ := strings.HasSuffix(text, "?")
            if (longEnough || endsWithQ) && rand.Float64() < 0.05 {
                // enforce daily auto-ask cap per chat (25)
                if _, allowed, err := c.mdb.TryConsumeAutoAskQuota(chatID, 25); err != nil || !allowed {
                    return
                }
                key := fmt.Sprintf("%d:%d", chatID, update.Message.ID)
                if !autoAskSent.LoadOrStore(key) {
                    userMsgs := []openAIMessage{{Role: "user", Content: text}}
                    client := &http.Client{Timeout: 60 * time.Second}
                    if answer, err := chatCompleteWithContinuation(ctx, client, userMsgs, c.GetSystemPrompt()); err == nil {
                        sendChunkedMessage(ctx, b, update.Message.Chat.ID, update.Message.ID, answer)
                    }
                }
            }
        }
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
