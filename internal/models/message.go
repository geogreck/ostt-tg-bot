package models

import "time"

type Message struct {
	ID                  int64
	ChatID              int64
	UserID              int64
	SentAt              time.Time
	ReactionCount       int
	ReactionBananaCount int
	ReactionLikeCount   int
	UserNickname        string
}

type MessageForSticker struct {
	UserNickname string
	Text         string
}
