package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
    defaultModel                 = "gpt://b1gslpsolbjjb92qq3b8/gpt-oss-20b/latest"
    chatCompletionsEndpoint      = "https://llm.api.cloud.yandex.net/v1/chat/completions"
)

type openAIMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type chatCompletionRequest struct {
    Model       string          `json:"model"`
    Messages    []openAIMessage `json:"messages"`
    Temperature float64         `json:"temperature,omitempty"`
    MaxTokens   int             `json:"max_tokens,omitempty"`
    Stream      bool            `json:"stream,omitempty"`
}

type chatCompletionResponse struct {
    Choices []struct {
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

func getAskContent(update *models.Update) ([]openAIMessage, error) {
    if update.Message == nil {
        return nil, errors.New("некорректное сообщение")
    }
    userMessages := []openAIMessage{}
    // 1) Если команда вызвана ответом на сообщение — добавим текст ответа
    if update.Message.ReplyToMessage != nil {
        replied := update.Message.ReplyToMessage
        if replied.Text != "" {
            userMessages = append(userMessages, openAIMessage{Role: "user", Content: replied.Text})
        } else if replied.Caption != "" {
            userMessages = append(userMessages, openAIMessage{Role: "user", Content: replied.Caption})
        }
    }
    // 2) Также пробуем вытащить инлайн-текст после "/ask"
    raw := strings.TrimSpace(update.Message.Text)
    if strings.HasPrefix(raw, "/ask") {
        if i := strings.IndexByte(raw, ' '); i != -1 {
            inline := strings.TrimSpace(raw[i+1:])
            if inline != "" {
                userMessages = append(userMessages, openAIMessage{Role: "user", Content: inline})
            }
        }
    }
    if len(userMessages) == 0 {
        return nil, errors.New("нужно ответить на текстовое сообщение или указать текст: /ask <сообщение>")
    }
    return userMessages, nil
}

func buildCompletionBody(userMessages []openAIMessage, systemPrompt string) ([]byte, error) {
    model := os.Getenv("OSST_AI_MODEL")
    if model == "" {
        model = os.Getenv("OSTT_AI_MODEL")
    }
    if model == "" {
        // legacy envs for compatibility
        if v := os.Getenv("OSST_AI_MODEL_URI"); v != "" {
            model = v
        }
    }
    if model == "" {
        model = defaultModel
    }

    messages := []openAIMessage{}
    if systemPrompt != "" {
        messages = append(messages, openAIMessage{Role: "system", Content: systemPrompt})
    }
    messages = append(messages, userMessages...)

    body := chatCompletionRequest{
        Model:       model,
        Messages:    messages,
        Temperature: 0.3,
        MaxTokens:   256,
        Stream:      false,
    }
    return json.Marshal(body)
}

func sendChatCompletion(ctx context.Context, client *http.Client, body []byte) (string, error) {
    apiKey := os.Getenv("OSST_AI_API_KEY")
    if apiKey == "" {
        apiKey = os.Getenv("OSTT_AI_API_KEY")
    }
    if apiKey == "" {
        return "", errors.New("переменная окружения OSST_AI_API_KEY не установлена")
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatCompletionsEndpoint, bytes.NewReader(body))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
    folderID := os.Getenv("OSST_AI_FOLDER_ID")
    if folderID == "" {
        folderID = os.Getenv("OSTT_AI_FOLDER_ID")
    }
    if folderID != "" {
        req.Header.Set("OpenAI-Project", folderID)
    }
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        b, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("ошибка запроса: %s: %s", resp.Status, string(b))
    }
    var out chatCompletionResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return "", err
    }
    if len(out.Choices) == 0 {
        return "", errors.New("пустой ответ модели")
    }
    return out.Choices[0].Message.Content, nil
}

// no polling required with Chat Completions

func (c *Commander) AskHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    // Admins: can use /ask anywhere, no limits
    userID := update.Message.From.ID
    isAdmin := userID == 354806980 || userID == 228020962

    // Chat access control: allow only a specific chat (supports supergroup negative IDs)
    const allowedChatShortID int64 = 2115621645
    chatID := update.Message.Chat.ID
    isAllowed := chatID == allowedChatShortID || (-chatID-1000000000000) == allowedChatShortID
    if !isAdmin && !isAllowed {
        b.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Команда /ask недоступна в этом чате.",
            ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
        })
        return
    }

    // Per-user daily quota enforcement
    if !isAdmin {
        if limitStr := os.Getenv("OSST_AI_ASK_DAILY_LIMIT"); limitStr != "" {
        var limit int
        _, err := fmt.Sscanf(limitStr, "%d", &limit)
        if err == nil && limit > 0 {
            remaining, allowed, qerr := c.mdb.TryConsumeAskQuota(update.Message.From.ID, limit)
            if qerr != nil {
                b.SendMessage(ctx, &bot.SendMessageParams{
                    ChatID: update.Message.Chat.ID,
                    Text:   fmt.Sprintf("Не удалось проверить квоту: %v", qerr),
                    ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
                })
                return
            }
            if !allowed {
                b.SendMessage(ctx, &bot.SendMessageParams{
                    ChatID: update.Message.Chat.ID,
                    Text:   fmt.Sprintf("Достигнут дневной лимит /ask. Попробуйте завтра. Осталось: %d", remaining),
                    ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
                })
                return
            }
        }
        }
    }

    userMsgs, err := getAskContent(update)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
		})
		return
	}
    body, err := buildCompletionBody(userMsgs, c.GetSystemPrompt())
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Не удалось подготовить запрос: %v", err),
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
		})
		return
	}
	client := &http.Client{Timeout: 60 * time.Second}
    answer, err := sendChatCompletion(ctx, client, body)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
            Text:   fmt.Sprintf("Ошибка запроса к модели: %v", err),
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   answer,
		ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
	})
}



