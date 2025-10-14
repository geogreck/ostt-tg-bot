package commands

import (
    "context"
    "strings"

    "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

func (c *Commander) ChangePromptHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    // Admin check
    userID := update.Message.From.ID
    isAdmin := userID == 354806980 || userID == 228020962
    if !isAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Недостаточно прав для изменения системного промпта.",
            ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
        })
        return
    }

    text := strings.TrimSpace(update.Message.Text)
    parts := strings.SplitN(text, " ", 2)
    if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Использование: /change_prompt <новый промпт>",
            ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
        })
        return
    }

    newPrompt := strings.TrimSpace(parts[1])
    c.SetSystemPrompt(newPrompt)

    b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "Системный промпт обновлён.",
        ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
    })
}

func (c *Commander) ShowPromptHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    // Admin check
    userID := update.Message.From.ID
    isAdmin := userID == 354806980 || userID == 228020962
    if !isAdmin {
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Недостаточно прав для просмотра системного промпта.",
            ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
        })
        return
    }

    current := c.GetSystemPrompt()
    if strings.TrimSpace(current) == "" {
        current = "(пусто)"
    }
    b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   current,
        ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
    })
}


