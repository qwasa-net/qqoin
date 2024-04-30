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

var userTableSQL = `
CREATE TABLE IF NOT EXISTS users (
	uid INTEGER PRIMARY KEY,
	username TEXT,
	name TEXT,
	data BLOB,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`

func (s *QStorage) UsersMigrate() {
	s.db.Exec(userTableSQL)
}

func (s *QStorage) GetUser(uid int64) (*User, error) {
	var user User
	row := s.db.QueryRow(`
	SELECT uid, username, name, created_at, updated_at
	FROM users WHERE uid = ?
	`, uid)
	err := row.Scan(&user.UID, &user.Username, &user.Name, &user.Created_at, &user.Updated_at)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *QStorage) CreateUser(user *User) error {
	var now = time.Now()
	_, err := s.db.Exec(`
	INSERT INTO users (uid, username, name, data, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	`, user.UID, user.Username, user.Name, user.Data, now, now)
	return err
}

func (s *QStorage) CreateUpdateUser(user *User) error {
	var now = time.Now()
	_, err := s.db.Exec(`
	INSERT INTO users (uid, username, name, data, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(uid) DO UPDATE SET username = ?, name = ?, updated_at = ?
	`, user.UID, user.Username, user.Name, user.Data, now, now, user.Username, user.Name, now)
	return err
}
