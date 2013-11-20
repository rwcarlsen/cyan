package query

import (
	"bytes"
	"database/sql"

	"github.com/rwcarlsen/cyan/nuc"
)

// SimIds returns a list of all simulation ids in the cyclus database for
// conn.
func SimIds(db *sql.DB) (ids []string, err error) {
	sql := "SELECT SimID FROM SimulationTimeInfo"
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var s string
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

type SimInfo struct {
	StartTime   int
	Duration    int
	DecayPeriod int
}

func SimStat(db *sql.DB, simid string) (si SimInfo, err error) {
	sql := "SELECT SimulationStart,Duration FROM SimulationTimeInfo WHERE SimID = ?"
	rows, err := db.Query(sql, simid)
	if err != nil {
		return si, err
	}

	for rows.Next() {
		if err := rows.Scan(&si.StartTime, &si.Duration); err != nil {
			return si, err
		}
	}
	if err := rows.Err(); err != nil {
		return si, err
	}
	return si, nil
}

func AllMatCreated(db *sql.DB, simid string) (m nuc.Material, err error) {
	sql := `SELECT cmp.IsoID,SUM(cmp.Quantity * res.Quantity) FROM (
				Resources As res
				INNER JOIN Compositions AS cmp ON res.StateID = cmp.ID
				INNER JOIN ResCreators AS cre ON res.ID = cre.ResID
			) WHERE (
				cre.SimID = ? AND cre.SimID = res.SimID AND cre.SimID = cmp.SimID
			) GROUP BY cmp.IsoID;`
	return makeMaterial(db, sql, simid)
}

// AllMatAt returns the simulation-global material inventory for the specified
// sim id at time t. Use t=-1 to specify end-of-simulation.
func AllMatAt(db *sql.DB, simid string, t int) (m nuc.Material, err error) {
	if t == -1 {
		si, err := SimStat(db, simid)
		if err != nil {
			return nil, err
		}
		t = si.StartTime + si.Duration
	}
	sql := `SELECT cmp.IsoID,SUM(cmp.Quantity * res.Quantity) FROM (
				Resources AS res
				INNER JOIN Compositions AS cmp ON cmp.ID = res.StateID
				INNER JOIN Inventories AS inv ON inv.ResID = res.ID
			) WHERE (
				inv.SimID = ? AND inv.SimID = res.SimID AND res.SimID = cmp.SimID
				AND inv.StartTime <= ? AND inv.EndTime > ?
			) GROUP BY cmp.IsoID;`
	return makeMaterial(db, sql, simid, t, t)
}

// Index builds an sql statement for creating a new index on the specified
// table over cols.  The index is named according to the table and cols.
func Index(table string, cols ...string) string {
	var buf bytes.Buffer
	buf.WriteString("CREATE INDEX IF NOT EXISTS ")
	buf.WriteString(table + "_" + cols[0])
	for _, c := range cols[1:] {
		buf.WriteString("_" + c)
	}
	buf.WriteString(" ON " + table + " (" + cols[0] + " ASC")
	for _, c := range cols[1:] {
		buf.WriteString("," + c + " ASC")
	}
	buf.WriteString(");")
	return buf.String()
}

func makeMaterial(db *sql.DB, sql string, args ...interface{}) (m nuc.Material, err error) {
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	m = nuc.Material{}
	var iso int
	var qty float64
	for rows.Next() {
		if err := rows.Scan(&iso, &qty); err != nil {
			return nil, err
		}
		m[nuc.Nuc(iso)] = nuc.Mass(qty)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return m, nil
}
