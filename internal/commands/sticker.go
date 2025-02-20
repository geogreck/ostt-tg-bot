package commands

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	imagegenerator "telegram-sticker-bot/internal/image-generator"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func setByChatId(chatId int64) string {
	if chatId == -1002115621645 || chatId == -1002307500811 {
		return "ostt_by_halaji_bot"
	}
	return ""
}

func CreateStickerSetHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	fmt.Println(strings.Split(update.Message.Text, " ")[1])
	fmt.Println("ID: ", update.Message.From.ID)

	name := strings.Split(update.Message.Text, " ")[1] + "_by_" + "halaji_bot"
	botInfo, _ := b.GetMe(ctx)

	imgBytes, err := imagegenerator.CreateSticker(name)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при создании набора стикеров: " + err.Error(),
		})
		return
	}
	stickerFile, _ := b.UploadStickerFile(ctx, &bot.UploadStickerFileParams{
		UserID: botInfo.ID,
		PngSticker: &models.InputFileUpload{
			Filename: "sticker.webp",
			Data:     bytes.NewReader(imgBytes),
		},
	})
	_, err = b.CreateNewStickerSet(ctx, &bot.CreateNewStickerSetParams{
		UserID: update.Message.From.ID,
		Name:   name,
		Title:  "dsadsa",
		Stickers: []models.InputSticker{
			{
				Sticker: &models.InputFileString{
					Data: stickerFile.FileID,
				},
				EmojiList: []string{"🍌"},
				Format:    "static",
				MaskPosition: models.MaskPosition{
					Point: "eyes",
				},
			},
		},
	})
	if err != nil {
		fmt.Println("ошибка при создании стикер пака:", err)
	}
}

type MultiStickerData struct {
	Message   string
	MessageId int
}

var UsersForMultiSticker sync.Map
var MultiStickerSemaphore sync.Mutex

func prepareAndSendSticker(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	imgBytes, err := imagegenerator.CreateSticker(text)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при создании стикера: " + err.Error(),
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})
		return
	}
	botInfo, _ := b.GetMe(ctx)
	stickerFile, err := b.UploadStickerFile(ctx, &bot.UploadStickerFileParams{
		UserID: botInfo.ID,
		PngSticker: &models.InputFileUpload{
			Filename: "sticker.webp",
			Data:     bytes.NewReader(imgBytes),
		},
	})
	if err != nil {
		fmt.Println("Ошибка при загрузке файла: ", err)
		return
	}

	if setByChatId(update.Message.Chat.ID) != "" {
		setId := setByChatId(update.Message.Chat.ID)
		res, err := b.AddStickerToSet(ctx, &bot.AddStickerToSetParams{
			UserID: botInfo.ID,
			Name:   setId,
			Sticker: models.InputSticker{
				Sticker: &models.InputFileString{
					Data: stickerFile.FileID,
				},
				EmojiList: []string{"🍌"},
				Format:    "static",
				MaskPosition: models.MaskPosition{
					Point: "eyes",
				},
			},
		})
		if err != nil || !res {
			fmt.Println("Ошибка при добавлении стикера в пак: ", err)
			return
		}

		set, err := b.GetStickerSet(ctx, &bot.GetStickerSetParams{
			Name: setId,
		})
		if err != nil || !res {
			fmt.Println("Ошибка при поиска пака: ", err)
			return
		}

		b.SendSticker(ctx, &bot.SendStickerParams{
			ChatID: update.Message.Chat.ID,
			Sticker: &models.InputFileString{
				Data: set.Stickers[len(set.Stickers)-1].FileID,
			},
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
			},
		})

	} else {
		b.SendSticker(ctx, &bot.SendStickerParams{
			ChatID: update.Message.Chat.ID,
			Sticker: &models.InputFileUpload{
				Filename: "sticker.webp",
				Data:     bytes.NewReader(imgBytes),
			},
		})
	}
}

func AddStickerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Text == "" {
		userId := update.Message.From.ID
		if messagesRaw, ok := UsersForMultiSticker.Load(userId); ok {
			messages := messagesRaw.([]MultiStickerData)
			messages = append(messages, MultiStickerData{
				update.Message.Text,
				update.Message.ID,
			})
			UsersForMultiSticker.Store(userId, messages)
		}
		go func() {
			MultiStickerSemaphore.Lock()
			UsersForMultiSticker.Store(userId, []MultiStickerData{})
			MultiStickerSemaphore.Unlock()

			time.Sleep(time.Second * 15)

			MultiStickerSemaphore.Lock()
			defer MultiStickerSemaphore.Unlock()

			if messagesRaw, ok := UsersForMultiSticker.Load(userId); !ok || len(messagesRaw.([]MultiStickerData)) == 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Ответьте на текстовое сообщение, чтобы создать стикер.",
					ReplyParameters: &models.ReplyParameters{
						MessageID: update.Message.ID,
					},
				})
				return
			}
			messagesRaw, _ := UsersForMultiSticker.Load(userId)
			messages := messagesRaw.([]MultiStickerData)
			sort.Slice(messages, func(i, j int) bool {
				return messages[i].MessageId < messages[j].MessageId
			})

			messageString := []string{}
			for _, message := range messages {
				messageString = append(messageString, message.Message)
			}

			prepareAndSendSticker(ctx, b, update, strings.Join(messageString, "\n"))
			UsersForMultiSticker.Delete(userId)
		}()

		return
	}

	prepareAndSendSticker(ctx, b, update, update.Message.ReplyToMessage.Text)
}
