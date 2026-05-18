CREATE TABLE IF NOT EXISTS complaints (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    city      TEXT    NOT NULL,
    street    TEXT    NOT NULL,
    latitude  REAL    NOT NULL,
    longitude REAL    NOT NULL,
    photo_url TEXT    NOT NULL
);