ALTER TABLE tg_messages
ADD COLUMN reaction_banana_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN reaction_like_count INTEGER NOT NULL DEFAULT 0;
