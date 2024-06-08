package storage

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

type QSOptions struct {
	StoragePath   string
	StorageEngine string
}

type QStorage struct {
	Opts     *QSOptions
	db       *sql.DB
	lock     sync.RWMutex
	prepared map[int]*sql.Stmt
}

func (s *QStorage) Open() (*sql.DB, error) {
	s.prepared = make(map[int]*sql.Stmt)
	pool, err := sql.Open(s.Opts.StorageEngine, s.Opts.StoragePath)
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
