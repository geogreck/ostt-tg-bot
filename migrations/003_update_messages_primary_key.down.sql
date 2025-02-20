ALTER TABLE tg_messages DROP CONSTRAINT IF EXISTS tg_messages_pkey;

ALTER TABLE tg_messages ADD PRIMARY KEY (message_id);
