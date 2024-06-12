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

var tapTableDDL = `
CREATE TABLE IF NOT EXISTS taps (
	uid INTEGER PRIMARY KEY,
	score INTEGER,
	count INTEGER,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`

func (s *QStorage) MigrateTaps() {
	s.db.Exec(tapTableDDL)
}

const tapGetStmt = 0
const tapUpsertStmt = 1
const tapInsertStmt = 2
const tapGetAllStmt = 9

func (s *QStorage) PrepareTaps() {

	//
	sqlGet := `
	SELECT uid, score, count, created_at, updated_at FROM taps WHERE uid=?
	`
	s.prepared[tapGetStmt], _ = s.db.Prepare(sqlGet)

	//
	sqlUpsert := `
	INSERT INTO taps (uid, score, count, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(uid) DO UPDATE SET count=count+1, score=score+?, updated_at=?
	`
	s.prepared[tapUpsertStmt], _ = s.db.Prepare(sqlUpsert)

	//
	sqlInsert := `
	INSERT INTO taps (uid, score, count, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)
	`
	s.prepared[tapInsertStmt], _ = s.db.Prepare(sqlInsert)

	//
	sqlGetAll := `
	SELECT uid, score, count, created_at, updated_at FROM taps ORDER BY count DESC LIMIT ?
	`
	s.prepared[tapGetAllStmt], _ = s.db.Prepare(sqlGetAll)

}

func (s *QStorage) GetTap(uid int64) (*Tap, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	tap := Tap{}
	row := s.prepared[tapGetStmt].QueryRow(uid)
	err := row.Scan(&tap.UID, &tap.Score, &tap.Count, &tap.Created_at, &tap.Updated_at)
	if err != nil {
		return nil, err
	}
	tap.ScoreTotal = tap.Score
	return &tap, nil
}

func (s *QStorage) CreateTap(tap *Tap) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	var now = time.Now()
	_, err := s.prepared[tapInsertStmt].Exec(tap.UID, tap.Energy, tap.Count, now, now)
	return err
}

func (s *QStorage) CreateUpdateTap(tap *Tap) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	var now = time.Now()
	_, err := s.prepared[tapUpsertStmt].Exec(tap.UID, tap.Energy, tap.Count, now, now, tap.Energy, now)
	return err
}

func (s *QStorage) GetAllTaps(limit int) ([]Tap, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	rows, err := s.prepared[tapGetAllStmt].Query(limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var taps []Tap
	for rows.Next() {
		var tap Tap
		err = rows.Scan(&tap.UID, &tap.Score, &tap.Count, &tap.Created_at, &tap.Updated_at)
		if err != nil {
			return nil, err
		}
		tap.ScoreTotal = tap.Score
		taps = append(taps, tap)
	}
	return taps, nil
}
