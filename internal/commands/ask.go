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
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	defaultYandexModelURI   = "gpt://b1gslpsolbjjb92qq3b8/yandexgpt/latest"
	completionAsyncEndpoint = "https://llm.api.cloud.yandex.net/foundationModels/v1/completionAsync"
	operationGetEndpoint    = "https://operation.api.cloud.yandex.net/operations/"
)

type yandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type yandexCompletionOptions struct {
	Stream           bool    `json:"stream"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        string  `json:"maxTokens"`
}

type yandexCompletionRequest struct {
	ModelURI          string                   `json:"modelUri"`
	CompletionOptions yandexCompletionOptions  `json:"completionOptions"`
	Messages          []yandexMessage          `json:"messages"`
}

type operationCreateResponse struct {
	ID string `json:"id"`
	Done bool  `json:"done"`
}

type operationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type operationGetResponse struct {
	Done     bool              `json:"done"`
	ID       string            `json:"id"`
	Error    *operationError   `json:"error"`
	Response *struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
	} `json:"response"`
}

func getReplyContent(update *models.Update) (string, error) {
	if update.Message == nil || update.Message.ReplyToMessage == nil {
		return "", errors.New("нужно ответить на сообщение командой /ask")
	}
	replied := update.Message.ReplyToMessage
	if replied.Text != "" {
		return replied.Text, nil
	}
	if replied.Caption != "" {
		return replied.Caption, nil
	}
	return "", errors.New("сообщение не содержит текст. Ответьте на текстовое сообщение")
}

func buildCompletionBody(text string) ([]byte, error) {
	modelURI := os.Getenv("OSTT_AI_MODEL_URI")
	if modelURI == "" {
		modelURI = defaultYandexModelURI
	}
    messages := []yandexMessage{}
    if systemPrompt := os.Getenv("OSTT_AI_SYSTEM_PROMPT"); systemPrompt != "" {
        messages = append(messages, yandexMessage{Role: "system", Text: systemPrompt})
    }
    messages = append(messages, yandexMessage{Role: "user", Text: text})

    body := yandexCompletionRequest{
        ModelURI: modelURI,
        CompletionOptions: yandexCompletionOptions{
            Stream:      false,
            Temperature: 0.3,
            MaxTokens:   "256",
        },
        Messages: messages,
    }
	return json.Marshal(body)
}

func doCompletionAsync(ctx context.Context, client *http.Client, body []byte) (string, error) {
	apiKey := os.Getenv("OSTT_AI_API_KEY")
	if apiKey == "" {
		return "", errors.New("переменная окружения OSTT_AI_API_KEY не установлена")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, completionAsyncEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", apiKey))
	if folderID := os.Getenv("OSTT_AI_FOLDER_ID"); folderID != "" {
		req.Header.Set("x-folder-id", folderID)
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
	var op operationCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&op); err != nil {
		return "", err
	}
	if op.ID == "" {
		return "", errors.New("пустой идентификатор операции")
	}
	return op.ID, nil
}

func pollOperationResult(ctx context.Context, client *http.Client, opID string) (string, error) {
	apiKey := os.Getenv("OSTT_AI_API_KEY")
	if apiKey == "" {
		return "", errors.New("переменная окружения OSTT_AI_API_KEY не установлена")
	}
	deadline := time.Now().Add(45 * time.Second)
	for attempt := 0; ; attempt++ {
		if time.Now().After(deadline) {
			return "", errors.New("время ожидания генерации истекло")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, operationGetEndpoint+opID, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", apiKey))
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		var op operationGetResponse
		func() {
			defer resp.Body.Close()
			_ = json.NewDecoder(resp.Body).Decode(&op)
		}()
		if op.Error != nil && op.Error.Message != "" {
			return "", fmt.Errorf("ошибка операции: %s", op.Error.Message)
		}
		if op.Done {
			if op.Response == nil || len(op.Response.Alternatives) == 0 {
				return "", errors.New("пустой ответ модели")
			}
			return op.Response.Alternatives[0].Message.Text, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *Commander) AskHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    // Per-user daily quota enforcement
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

	content, err := getReplyContent(update)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   err.Error(),
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
		})
		return
	}
	body, err := buildCompletionBody(content)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Не удалось подготовить запрос: %v", err),
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
		})
		return
	}
	client := &http.Client{Timeout: 60 * time.Second}
	opID, err := doCompletionAsync(ctx, client, body)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Ошибка запроса к модели: %v", err),
			ReplyParameters: &models.ReplyParameters{MessageID: update.Message.ID},
		})
		return
	}
	answer, err := pollOperationResult(ctx, client, opID)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Не удалось получить результат: %v", err),
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


