package query

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/rwcarlsen/cyan/nuc"
)

// SimIds returns a list of all simulation ids in the cyclus database for
// conn.
func SimIds(db *sql.DB) (ids [][]byte, err error) {
	sql := "SELECT SimId FROM Info"
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

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

type SimInfo struct {
	Id          []byte
	Duration    int
	DecayPeriod int
}

func (si SimInfo) String() string {
	return fmt.Sprintf("%x: dur=%v, decay=%v", si.Id,
		si.Duration, si.DecayPeriod)
}

func SimStat(db *sql.DB, simid []byte) (si SimInfo, err error) {
	sql := "SELECT Duration,DecayInterval FROM Info WHERE SimId = ?"
	rows, err := db.Query(sql, simid)
	if err != nil {
		return si, err
	}
	for rows.Next() {
		if err := rows.Scan(&si.Duration, &si.DecayPeriod); err != nil {
			return si, err
		}
	}
	if err := rows.Err(); err != nil {
		return si, err
	}

	si.Id = simid
	return si, nil
}

type AgentInfo struct {
	Id       int
	Kind     string
	Impl     string
	Proto    string
	Parent   int
	Lifetime int
	Enter    int
	Exit     int
}

func (ai AgentInfo) String() string {
	return fmt.Sprintf("%v %v:%v:%v: parent=%v, lifetime=%v, enter=%v, exit=%v", ai.Id,
		ai.Kind, ai.Impl, ai.Proto, ai.Parent, ai.Lifetime, ai.Enter, ai.Exit)
}

func AllAgents(db *sql.DB, simid []byte, proto string) (ags []AgentInfo, err error) {
	s := `SELECT AgentId,Kind,Implementation,Prototype,ParentId,EnterTime,ExitTime,Lifetime FROM
				Agents
			WHERE Agents.SimId = ?`

	var rows *sql.Rows
	if proto != "" {
		s += ` AND Agents.Prototype = ?`
		rows, err = db.Query(s, simid, proto)
	} else {
		rows, err = db.Query(s, simid)
	}
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		ai := AgentInfo{}
		var exit sql.NullInt64
		if err := rows.Scan(&ai.Id, &ai.Kind, &ai.Impl, &ai.Proto, &ai.Parent, &ai.Enter, &exit, &ai.Lifetime); err != nil {
			return nil, err
		}
		if !exit.Valid {
			exit.Int64 = -1
		}
		ai.Exit = int(exit.Int64)
		ags = append(ags, ai)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ags, nil
}

func DeployCumulative(db *sql.DB, simid []byte, proto string) (xys []XY, err error) {
	sql := `SELECT ti.Time,COUNT(*)
			  FROM TimeList AS ti 
			  LEFT JOIN Agents AS ag ON ti.Time >= ag.EnterTime AND (ag.ExitTime > ti.Time OR ag.ExitTime IS NULL)
			WHERE
			  ti.SimId = ag.SimId
			  AND ag.SimId = ?
			  AND ag.Prototype = ?
			GROUP BY ti.Time
			ORDER BY ti.Time;`
	rows, err := db.Query(sql, simid, proto)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		xy := XY{}
		if err := rows.Scan(&xy.X, &xy.Y); err != nil {
			return nil, err
		}
		xys = append(xys, xy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return xys, nil
}

type XY struct {
	X int
	Y float64
}

func InvSeries(db *sql.DB, simid []byte, agent int, iso int) (xys []XY, err error) {
	sql := `SELECT ti.Time,SUM(cmp.MassFrac * inv.Quantity) FROM (
				Compositions AS cmp
				INNER JOIN Inventories AS inv ON inv.StateId = cmp.StateId
				INNER JOIN TimeList AS ti ON (ti.Time >= inv.StartTime AND ti.Time < inv.EndTime)
			) WHERE (
				inv.SimId = ? AND inv.SimId = cmp.SimId AND ti.SimId = inv.SimId
				AND inv.AgentId = ? AND cmp.NucId = ?
			) GROUP BY ti.Time,cmp.NucId;`
	rows, err := db.Query(sql, simid, agent, iso)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		xy := XY{}
		if err := rows.Scan(&xy.X, &xy.Y); err != nil {
			return nil, err
		}
		xys = append(xys, xy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return xys, nil
}

// MatCreated returns the total amount of material created by the listed
// agent ids in the simulation for the given sim id between t0 and t1. Passing no
// agents defaults to all agents. Use t0=0 to specify beginning-of-simulation.
// Use t1=-1 to specify end-of-simulation.
func MatCreated(db *sql.DB, simid []byte, t0, t1 int, agents ...int) (m nuc.Material, err error) {
	if t1 == -1 {
		si, err := SimStat(db, simid)
		if err != nil {
			return nil, err
		}
		t1 = si.Duration
	}
	filt := ""
	if len(agents) > 0 {
		filt += " AND cre.AgentId IN (" + strconv.Itoa(agents[0])
		for _, a := range agents[1:] {
			filt += "," + strconv.Itoa(a)
		}
		filt += ") "
	}

	sql := `SELECT cmp.NucId,SUM(cmp.MassFrac * res.Quantity) FROM (
				Resources As res
				INNER JOIN Compositions AS cmp ON res.StateId = cmp.StateId
				INNER JOIN ResCreators AS cre ON res.ResourceId = cre.ResourceId
			) WHERE (
				cre.SimId = ? AND cre.SimId = res.SimId AND cre.SimId = cmp.SimId
				AND res.TimeCreated >= ? AND res.TimeCreated < ?`
	sql += filt
	sql += `) GROUP BY cmp.NucId;`
	return makeMaterial(db, sql, simid, t0, t1)
}

// InvAt returns the material inventory of the listed agent ids for the
// specified sim id at time t. Passing no agents defaults to all agents. Use
// t=-1 to specify end-of-simulation.
func InvAt(db *sql.DB, simid []byte, t int, agents ...int) (m nuc.Material, err error) {
	if t == -1 {
		si, err := SimStat(db, simid)
		if err != nil {
			return nil, err
		}
		t = si.Duration
	}
	filt := ""
	if len(agents) > 0 {
		filt += " AND inv.AgentId IN (" + strconv.Itoa(agents[0])
		for _, a := range agents[1:] {
			filt += "," + strconv.Itoa(a)
		}
		filt += ") "
	}
	sql := `SELECT cmp.NucId,SUM(cmp.MassFrac * inv.Quantity) FROM (
				Inventories AS inv
				INNER JOIN Compositions AS cmp ON inv.StateId = cmp.StateId
			) WHERE (
				inv.SimId = ? AND inv.SimId = cmp.SimId
				AND inv.StartTime <= ? AND inv.EndTime > ?`
	sql += filt
	sql += `) GROUP BY cmp.NucId;`
	return makeMaterial(db, sql, simid, t, t)
}

// InvMassAt returns the mass of material inventory of the listed agent ids
// for the specified sim id at time t. Passing no agents defaults to all
// agents. Use t=-1 to specify end-of-simulation.
func InvMassAt(db *sql.DB, simid []byte, t int, agents ...int) (mass float64, err error) {
	m, err := InvAt(db, simid, t, agents...)
	if err != nil {
		return 0, err
	}
	return float64(m.Mass()), nil
}

type FlowArc struct {
	Src      string
	Dst      string
	Commod   string
	Quantity float64
}

func FlowGraph(db *sql.DB, simid []byte, t0, t1 int, groupByProto bool) (arcs []FlowArc, err error) {
	if t1 == -1 {
		si, err := SimStat(db, simid)
		if err != nil {
			return nil, err
		}
		t1 = si.Duration
	}

	var sql string
	if !groupByProto {
		sql = `SELECT snd.Prototype || " " || tr.SenderId,rcv.Prototype || " " || tr.ReceiverId,tr.Commodity,SUM(res.Quantity) FROM (
					Resources AS res
					INNER JOIN Transactions AS tr ON tr.ResourceId = res.ResourceId
					INNER JOIN Agents AS snd ON snd.AgentId = tr.SenderId
					INNER JOIN Agents AS rcv ON rcv.AgentId = tr.ReceiverId
				) WHERE (
					res.SimId = ? AND tr.SimId = res.SimId
					AND tr.Time >= ? AND tr.Time < ?
				) GROUP BY tr.SenderId,tr.ReceiverId,tr.Commodity;`
	} else {
		sql = `SELECT snd.Prototype,rcv.Prototype,tr.Commodity,SUM(res.Quantity) FROM (
					Resources AS res
					INNER JOIN Transactions AS tr ON tr.ResourceId = res.ResourceId
					INNER JOIN Agents AS snd ON snd.AgentId = tr.SenderId
					INNER JOIN Agents AS rcv ON rcv.AgentId = tr.ReceiverId
				) WHERE (
					res.SimId = ? AND tr.SimId = res.SimId
					AND tr.Time >= ? AND tr.Time < ?
				) GROUP BY snd.Prototype,rcv.Prototype,tr.Commodity;`
	}

	rows, err := db.Query(sql, simid, t0, t1)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		arc := FlowArc{}
		if err := rows.Scan(&arc.Src, &arc.Dst, &arc.Commod, &arc.Quantity); err != nil {
			return nil, err
		}
		arcs = append(arcs, arc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return arcs, nil
}

func Flow(db *sql.DB, simid []byte, t0, t1 int, fromAgents, toAgents []int) (m nuc.Material, err error) {
	if t1 == -1 {
		si, err := SimStat(db, simid)
		if err != nil {
			return nil, err
		}
		t1 = si.Duration
	}
	filt := " AND tr.SenderId IN (" + strconv.Itoa(fromAgents[0])
	for _, a := range fromAgents[1:] {
		filt += "," + strconv.Itoa(a)
	}
	filt += ") "
	filt += " AND tr.ReceiverId IN (" + strconv.Itoa(toAgents[0])
	for _, a := range toAgents[1:] {
		filt += "," + strconv.Itoa(a)
	}
	filt += ") "

	sql := `SELECT cmp.NucId,SUM(cmp.MassFrac * res.Quantity) FROM (
				Resources AS res
				INNER JOIN Compositions AS cmp ON cmp.StateId = res.StateId
				INNER JOIN Transactions AS tr ON tr.ResourceId = res.ResourceId
			) WHERE (
				res.SimId = ? AND cmp.SimId = res.SimId AND tr.SimId = res.SimId
				AND tr.Time >= ? AND tr.Time < ?`
	sql += filt
	sql += `) GROUP BY cmp.NucId;`
	return makeMaterial(db, sql, simid, t0, t1)
}

// EnergyProduced returns the total amount of energy produced between t0 and
// t1 in Joules. Use t1=-1 to specify end-of-simulation.
func EnergyProduced(db *sql.DB, simid []byte, t0, t1 int) (float64, error) {
	t2 := t1 + 1
	if t1 < 0 {
		t2 = -1
	}
	created, err := MatCreated(db, simid, t0+1, t2)
	if err != nil {
		return 0, err
	}
	mat0, err := InvAt(db, simid, t0)
	if err != nil {
		return 0, err
	}
	mat1, err := InvAt(db, simid, t1)
	if err != nil {
		return 0, err
	}

	fpeCreated := nuc.FPE(created)
	fpe0 := nuc.FPE(mat0)
	fpe1 := nuc.FPE(mat1)

	return fpe0 - (fpe1 - fpeCreated), nil
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
