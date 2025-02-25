package pgsql

import (
	"database/sql"
	"fmt"
	"telegram-sticker-bot/internal/models"
)

type MessagesDatabase struct {
	db *sql.DB
}

func NewMessagesDatabase(db *sql.DB) MessagesDatabase {
	return MessagesDatabase{db: db}
}

func (md *MessagesDatabase) Save(message models.Message) error {
	query := `
		INSERT INTO tg_messages (chat_id, message_id, user_id, sent_at, author_nickname)
		VALUES ($1, $2, $3, $4, $5);
	`
	err := md.db.QueryRow(query, message.ChatID, message.ID, message.UserID, message.SentAt, message.UserNickname).Scan(&message.ID)
	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}
	return nil
}

func (md *MessagesDatabase) UpdateReactionCount(messageId int64, newCount int) error {
	query := `
		UPDATE tg_messages
		SET reaction_count = $1
		WHERE message_id = $2;
	`
	result, err := md.db.Exec(query, newCount, messageId)
	if err != nil {
		return fmt.Errorf("failed to update reaction count: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("no message updated with id %d", messageId)
	}

	return nil
}

func (md *MessagesDatabase) UpdateBananaReactionCount(messageId int, chatId int64, delta int) error {
	query := `
		UPDATE tg_messages
		SET reaction_banana_count = reaction_banana_count + $1
		WHERE chat_id = $2 AND message_id = $3;
	`
	result, err := md.db.Exec(query, delta, chatId, messageId)
	if err != nil {
		return fmt.Errorf("failed to update banana reaction count: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("no message updated with id %d", messageId)
	}
	return nil
}

func (md *MessagesDatabase) UpdateLikeReactionCount(messageId int, chatId int64, delta int) error {
	query := `
		UPDATE tg_messages
		SET reaction_like_count = reaction_like_count + $1
		WHERE chat_id = $2 AND message_id = $3;
	`
	result, err := md.db.Exec(query, delta, chatId, messageId)
	if err != nil {
		return fmt.Errorf("failed to update like reaction count: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve affected rows: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("no message updated with id %d", messageId)
	}
	return nil
}

func (md *MessagesDatabase) GetReactions(messageId int, chatId int64) (bananaCount int, likeCount int, err error) {
	query := `
		SELECT reaction_banana_count, reaction_like_count
		FROM tg_messages
		WHERE chat_id = $1 AND message_id = $2;
	`
	err = md.db.QueryRow(query, chatId, messageId).Scan(&bananaCount, &likeCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get reactions for message: %v", err)
	}
	return bananaCount, likeCount, nil
}

func (md *MessagesDatabase) GetTopMessages(chatId int64) ([]models.Message, error) {
	// timeStr := "1d"
	// timeParsed, err := time.ParseDuration(timeStr)
	// if err != nil {
	// 	return []models.Message{}, fmt.Errorf("Неверный формат даты")
	// }
	query := `
		SELECT message_id, reaction_banana_count, reaction_like_count, reaction_like_count+2*reaction_banana_count AS score, author_nickname, sent_at
		FROM tg_messages
		WHERE chat_id = $1
		ORDER BY score DESC
		LIMIT 20;
	`
	rows, err := md.db.Query(query, chatId)
	if err != nil {
		return []models.Message{}, fmt.Errorf("failed to get top messages: %v", err)
	}
	messages := make([]models.Message, 0)
	for rows.Next() {
		message := models.Message{
			ChatID: chatId,
		}
		score := 0
		err := rows.Scan(&message.ID, &message.ReactionBananaCount, &message.ReactionLikeCount, &score, &message.UserNickname, &message.SentAt)
		if err != nil {
			return []models.Message{}, fmt.Errorf("failed to scan top messages: %v", err)
		}
		messages = append(messages, message)
	}
	return messages, nil
}
