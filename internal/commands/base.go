package commands

import (
	"database/sql"
	"log"
	"telegram-sticker-bot/internal/db/pgsql"

	_ "github.com/lib/pq"
)

type Commander struct {
	mdb pgsql.MessagesDatabase
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
		mdb: pgsql.NewMessagesDatabase(db),
	}
}
