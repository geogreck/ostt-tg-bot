CREATE TABLE IF NOT EXISTS ask_usage (
    user_id BIGINT NOT NULL,
    usage_date DATE NOT NULL,
    used INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date)
);


