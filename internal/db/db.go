package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS wg_server (
    id          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    address     TEXT    NOT NULL,
    listen_port INTEGER NOT NULL,
    private_key TEXT    NOT NULL,
    public_key  TEXT    NOT NULL,
    mtu         INTEGER NOT NULL,
    dns         TEXT,
    post_up     TEXT,
    post_down   TEXT,
    endpoint    TEXT,
    comments    TEXT
);

CREATE TABLE IF NOT EXISTS sys_users (
    id     INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name   TEXT    NOT NULL,
    passwd TEXT    NOT NULL,
    roles  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS wg_clients (
    id                   INTEGER           NOT NULL PRIMARY KEY AUTOINCREMENT,
    server_id            INTEGER           NOT NULL,
    name                 TEXT              NOT NULL,
    address              TEXT              NOT NULL,
    listen_port          INTEGER,
    private_key          TEXT              NOT NULL,
    public_key           TEXT              NOT NULL,
    allow_ips            TEXT              NOT NULL,
    mtu                  INTEGER           NOT NULL,
    dns                  TEXT,
    description          TEXT,
    comments             TEXT,
    disabled             INTEGER DEFAULT 0 NOT NULL,
    persistent_keepalive INTEGER           NOT NULL
);
`)
	return err
}
