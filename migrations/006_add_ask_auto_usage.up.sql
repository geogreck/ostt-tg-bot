CREATE TABLE IF NOT EXISTS ask_auto_usage (
    chat_id BIGINT NOT NULL,
    usage_date DATE NOT NULL,
    used INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (chat_id, usage_date)
);


