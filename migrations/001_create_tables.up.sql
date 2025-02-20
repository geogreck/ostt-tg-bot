-- Таблица для хранения информации о чатах
CREATE TABLE tg_chats (
    chat_id BIGSERIAL PRIMARY KEY,
    chat_name VARCHAR(255) NOT NULL
);

-- Таблица для хранения информации об авторах (пользователях)
CREATE TABLE tg_users (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE,  -- можно хранить логин или никнейм
    first_name VARCHAR(100),
    last_name VARCHAR(100)
);

-- Таблица для хранения сообщений
CREATE TABLE tg_messages (
    message_id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL,      -- дата и время отправки (с учетом часового пояса)
    reaction_count INTEGER NOT NULL DEFAULT 0  -- количество реакций
);

