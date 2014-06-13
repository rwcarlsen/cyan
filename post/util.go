package post

import (
	"database/sql"
	"time"
)

// GetSimIds returns a list of all simulation ids in the cyclus database for
// conn.
func GetSimIds(db *sql.DB) (ids [][]byte, err error) {
	sql := "SELECT SimID FROM Info"
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s []byte
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		ids = append(ids, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func panicif(err error) {
	if err != nil {
		panic(err.Error())
	}
}

type Timer struct {
	starts map[string]time.Time
	Totals map[string]time.Duration
}

func NewTimer() *Timer {
	return &Timer{
		map[string]time.Time{},
		map[string]time.Duration{},
	}
}

func (t *Timer) Start(label string) {
	if _, ok := t.starts[label]; !ok {
		t.starts[label] = time.Now()
	}
}

func (t *Timer) Stop(label string) {
	if start, ok := t.starts[label]; ok {
		t.Totals[label] += time.Now().Sub(start)
	}
	delete(t.starts, label)
}

type NullWriter struct{}

func (_ NullWriter) Write(p []byte) (n int, err error) { return len(p), nil }
