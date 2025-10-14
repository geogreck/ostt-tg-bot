package commands

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"telegram-sticker-bot/internal/db/pgsql"

	_ "github.com/lib/pq"
)

type Commander struct {
	mdb           pgsql.MessagesDatabase
	systemPrompt  string
	promptRWMutex sync.RWMutex
}

func NewCommander() Commander {
	connStr := "postgres://postgres:postgres@db:5432/telegram_db?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	return Commander{
		mdb:          pgsql.NewMessagesDatabase(db),
		systemPrompt: os.Getenv("OSST_AI_SYSTEM_PROMPT"),
	}
}

func (c *Commander) GetSystemPrompt() string {
	c.promptRWMutex.RLock()
	defer c.promptRWMutex.RUnlock()
	return c.systemPrompt
}

func (c *Commander) SetSystemPrompt(p string) {
	c.promptRWMutex.Lock()
	defer c.promptRWMutex.Unlock()
	c.systemPrompt = p
}
