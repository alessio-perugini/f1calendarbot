CREATE TABLE subscribers (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    telegram_id INTEGER NOT NULL UNIQUE
);