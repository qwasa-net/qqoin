package storage

import (
	"database/sql"
	"log"
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

func NewQStorage(opts *QSOptions) *QStorage {
	storage := QStorage{
		Opts: opts,
	}
	storage.Open()
	storage.Migrate()
	storage.Prepare()
	return &storage
}

func (s *QStorage) Open() (*sql.DB, error) {
	s.prepared = make(map[int]*sql.Stmt)
	log.Printf("storage db file: %s:%s\n", s.Opts.StorageEngine, s.Opts.StoragePath)
	pool, err := sql.Open(s.Opts.StorageEngine, s.Opts.StoragePath)
	s.db = pool
	return pool, err
}

func (s *QStorage) Migrate() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.MigrateUsers()
	s.MigrateTaps()
	s.MigrateQqokens()
}

func (s *QStorage) Prepare() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.PrepareTaps()
	s.PrepareUsers()
	s.PrepareQqokens()
}

func (s *QStorage) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.db.Exec("VACUUM")
	log.Println("storage closing …")
	return s.db.Close()
}
