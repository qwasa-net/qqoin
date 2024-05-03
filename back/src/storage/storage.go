package storage

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

type QStorage struct {
	engine   string
	path     string
	db       *sql.DB
	lock     sync.RWMutex
	prepared map[int]*sql.Stmt
}

func (s *QStorage) Open(path string, engine string) (*sql.DB, error) {
	s.engine = "sqlite"
	s.path = path
	s.prepared = make(map[int]*sql.Stmt)
	pool, err := sql.Open(s.engine, path)
	s.db = pool
	return pool, err
}

func (s *QStorage) Migrate() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.MigrateUsers()
	s.MigrateTaps()
}

func (s *QStorage) Prepare() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.PrepareTaps()
	s.PrepareUsers()
}

func (s *QStorage) Close() error {
	return s.db.Close()
}
