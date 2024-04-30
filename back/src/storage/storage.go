package storage

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type QStorage struct {
	engine string
	path   string
	db     *sql.DB
}

func (s *QStorage) Open(path string, engine string) (*sql.DB, error) {
	s.engine = "sqlite"
	s.path = path
	pool, err := sql.Open(s.engine, path)
	s.db = pool
	return pool, err
}

func (s *QStorage) Migrate() {
	s.UsersMigrate()
	s.TapsMigrate()
}
