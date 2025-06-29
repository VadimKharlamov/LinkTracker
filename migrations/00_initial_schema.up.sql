CREATE TABLE IF NOT EXISTS chats (
    id bigint PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS links (
    id SERIAL PRIMARY KEY,
    link VARCHAR(255) NOT NULL,
    tags TEXT[],
    filters TEXT[],
    lastupdated TIMESTAMP NOT NULL,
    chatid bigint NOT NULL,
    FOREIGN KEY (chatid) REFERENCES chats(id) ON DELETE CASCADE
);

ALTER TABLE links ADD CONSTRAINT unique_link_chatid UNIQUE (link, chatId);

CREATE INDEX idx_links_chatid ON links(chatid);
CREATE INDEX idx_unique_link_chatid ON links(link, chatid);