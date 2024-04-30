package storage

import (
	"time"
)

type Tap struct {
	UID        int64     `json:"uid"`
	Score      int64     `json:"score"`
	ScoreTotal int64     `json:"score_total"`
	Count      int64     `json:"count"`
	Energy     int64     `json:"energy"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

var tapTableSQL = `
CREATE TABLE IF NOT EXISTS taps (
	uid INTEGER PRIMARY KEY,
	score INTEGER,
	count INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`

func (s *QStorage) TapsMigrate() {
	s.db.Exec(tapTableSQL)
}

func (s *QStorage) GetTap(uid int64) (*Tap, error) {
	var tap Tap
	row := s.db.QueryRow(`SELECT uid, score, count, created_at, updated_at FROM taps WHERE uid = ?`, uid)
	err := row.Scan(&tap.UID, &tap.Score, &tap.Count, &tap.Created_at, &tap.Updated_at)
	if err != nil {
		return nil, err
	}
	tap.ScoreTotal = tap.Score
	return &tap, nil
}

func (s *QStorage) CreateTap(tap *Tap) error {
	var now = time.Now()
	_, err := s.db.Exec(`
	INSERT INTO taps (uid, score, count, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	`, tap.UID, tap.Score, tap.Count, now, now)
	return err
}

func (s *QStorage) CreateUpdateTap(tap *Tap) error {
	var now = time.Now()
	_, err := s.db.Exec(`
	INSERT INTO taps (uid, score, count, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(uid) DO UPDATE SET score = score + ?, count = count+1, updated_at = ?
	`, tap.UID, tap.Score, tap.Count, now, now, tap.Energy, now)
	return err
}
