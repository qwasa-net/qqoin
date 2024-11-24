package storage

import (
	"time"
)

type QQoken struct {
	UID         int64     `json:"uid"`
	Wallet_addr string    `json:"wallet_addr"`
	Qqoken_addr string    `json:"qqoken_addr"`
	Qqoken_id   string    `json:"qqoken_id"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

var qqokenTableDDL = `
CREATE TABLE IF NOT EXISTS qqokens (
	uid INTEGER PRIMARY KEY,
	wallet_addr TEXT,
	qqoken_addr TEXT,
	qqoken_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`

func (s *QStorage) MigrateQqokens() {
	s.db.Exec(qqokenTableDDL)
}

const qqokenGetStmt = 21
const qqokenUpsertStmt = 22
const qqokenInsertStmt = 23
const qqokenGetAllStmt = 24

func (s *QStorage) PrepareQqokens() {
	//
	sqlGet := `
	SELECT uid, wallet_addr, qqoken_addr, qqoken_id, created_at, updated_at
	FROM qqokens WHERE uid=?
	`
	s.prepared[qqokenGetStmt], _ = s.db.Prepare(sqlGet)

	//
	sqlUpsert := `
	INSERT INTO qqokens (uid, wallet_addr, qqoken_addr, qqoken_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(uid) DO UPDATE SET wallet_addr=?, updated_at=?
	`
	s.prepared[qqokenUpsertStmt], _ = s.db.Prepare(sqlUpsert)

	//
	sqlInsert := `
	INSERT INTO qqokens (uid, wallet_addr, qqoken_addr, qqoken_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	s.prepared[qqokenInsertStmt], _ = s.db.Prepare(sqlInsert)

	//
	sqlGetAll := `
	SELECT uid, wallet_addr, qqoken_addr, qqoken_id, created_at, updated_at
	FROM qqokens ORDER BY created_at DESC LIMIT ?
	`
	s.prepared[qqokenGetAllStmt], _ = s.db.Prepare(sqlGetAll)

}

func (s *QStorage) GetQqoken(uid int64) (*QQoken, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var qqoken QQoken
	row := s.prepared[qqokenGetStmt].QueryRow(uid)
	err := row.Scan(
		&qqoken.UID,
		&qqoken.Wallet_addr,
		&qqoken.Qqoken_addr,
		&qqoken.Qqoken_id,
		&qqoken.Created_at,
		&qqoken.Updated_at)
	if err != nil {
		return nil, err
	}
	return &qqoken, nil
}

func (s *QStorage) CreateQqoken(qqoken *QQoken) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var now = time.Now()
	_, err := s.prepared[qqokenInsertStmt].Exec(
		qqoken.UID,
		qqoken.Wallet_addr,
		qqoken.Qqoken_addr,
		qqoken.Qqoken_id,
		now,
		now)
	return err
}

func (s *QStorage) CreateUpdateQqoken(qqoken *QQoken) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var now = time.Now()
	_, err := s.prepared[qqokenUpsertStmt].Exec(
		qqoken.UID,
		qqoken.Wallet_addr,
		qqoken.Qqoken_addr,
		qqoken.Qqoken_id,
		now,
		now,
		qqoken.Wallet_addr,
		now)
	return err
}

func (s *QStorage) GetAllQqokens(limit int) ([]QQoken, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rows, err := s.prepared[qqokenGetAllStmt].Query(limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var qqokens []QQoken
	for rows.Next() {
		var qqoken QQoken
		err = rows.Scan(
			&qqoken.UID,
			&qqoken.Wallet_addr,
			&qqoken.Qqoken_addr,
			&qqoken.Qqoken_id,
			&qqoken.Created_at,
			&qqoken.Updated_at)
		if err != nil {
			return nil, err
		}
		qqokens = append(qqokens, qqoken)
	}
	return qqokens, nil
}
