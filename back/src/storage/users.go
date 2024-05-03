package storage

import (
	"time"
)

type User struct {
	UID        int64     `json:"uid"`
	Username   string    `json:"username"`
	Name       string    `json:"name"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Data       *byte     `json:"data"`
}

var userTableDDL = `
CREATE TABLE IF NOT EXISTS users (
	uid INTEGER PRIMARY KEY,
	username TEXT,
	name TEXT,
	data BLOB,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`

func (s *QStorage) MigrateUsers() {
	s.db.Exec(userTableDDL)
}

const userGetStmt = 5
const userUpsertStmt = 6
const userInsertStmt = 7

func (s *QStorage) PrepareUsers() {
	//
	sqlGet := `
	SELECT uid, username, name, created_at, updated_at FROM users WHERE uid=?
	`
	s.prepared[userGetStmt], _ = s.db.Prepare(sqlGet)

	//
	sqlUpsert := `
	INSERT INTO users (uid, username, name, data, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(uid) DO UPDATE SET username=?, name=?, updated_at=?
	`
	s.prepared[userUpsertStmt], _ = s.db.Prepare(sqlUpsert)

	//
	sqlInsert := `
	INSERT INTO users (uid, username, name, data, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	`
	s.prepared[userInsertStmt], _ = s.db.Prepare(sqlInsert)

}

func (s *QStorage) GetUser(uid int64) (*User, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var user User
	row := s.prepared[userGetStmt].QueryRow(uid)
	err := row.Scan(&user.UID, &user.Username, &user.Name, &user.Created_at, &user.Updated_at)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *QStorage) CreateUser(user *User) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var now = time.Now()
	_, err := s.prepared[userInsertStmt].Exec(user.UID, user.Username, user.Name, user.Data, now, now)
	return err
}

func (s *QStorage) CreateUpdateUser(user *User) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var now = time.Now()
	_, err := s.prepared[userUpsertStmt].Exec(user.UID, user.Username, user.Name, user.Data, now, now, user.Username, user.Name, now)
	return err
}
