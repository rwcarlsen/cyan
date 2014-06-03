package post

import (
	"fmt"
	"io"
	"log"
	"math"

	"github.com/mxk/go-sqlite/sqlite3"
	"github.com/rwcarlsen/cyan/query"
)

// The number of sql commands to buffer before dumping to the output database.
const DumpFreq = 100000

var (
	preExecStmts = []string{
		"CREATE TABLE IF NOT EXISTS AgentExit (SimId BLOB,AgentId INTEGER,ExitTime INTEGER);",
		"CREATE TABLE IF NOT EXISTS Resources (SimId INTEGER,ResourceId INTEGER,ObjId INTEGER,Type TEXT,TimeCreated INTEGER,Quantity REAL,Units TEXT,QualId INTEGER,Parent1 INTEGER,Parent2 INTEGER);",
		"CREATE TABLE IF NOT EXISTS ResCreators (SimId INTEGER,ResourceId INTEGER,AgentId INTEGER);",
		"CREATE TABLE IF NOT EXISTS Agents (SimId BLOB,AgentId INTEGER,Kind TEXT,Spec TEXT,Prototype TEXT,ParentId INTEGER,Lifetime INTEGER,EnterTime INTEGER,ExitTime INTEGER);",
		"CREATE TABLE IF NOT EXISTS Inventories (SimId BLOB,ResourceId INTEGER,AgentId INTEGER,StartTime INTEGER,EndTime INTEGER,QualId INTEGER,Quantity REAL);",
		"CREATE TABLE IF NOT EXISTS TimeList (SimId BLOB, Time INTEGER);",
		"CREATE TABLE IF NOT EXISTS Transactions (SimId BLOB, TransactionId INTEGER, SenderId INTEGER, ReceiverId INTEGER, ResourceId INTEGER, Commodity TEXT, Time INTEGER);",
		query.Index("TimeList", "Time"),
		query.Index("Resources", "SimId", "ResourceId", "QualId"),
		query.Index("Compositions", "SimId", "QualId", "NucId"),
		query.Index("Transactions", "SimId", "ResourceId"),
		query.Index("Transactions", "TransactionId"),
		query.Index("ResCreators", "SimId", "ResourceId"),
	}
	postExecStmts = []string{
		query.Index("Agents", "SimId", "Prototype"),
		query.Index("Inventories", "SimId", "AgentId"),
		query.Index("Inventories", "SimId", "ResourceId", "StartTime"),
		"ANALYZE;",
	}
	dumpSql    = "INSERT INTO Inventories VALUES (?,?,?,?,?,?,?);"
	resSqlHead = "SELECT ResourceId,TimeCreated,QualId,Quantity FROM "
	resSqlTail = " WHERE Parent1 = ? OR Parent2 = ?;"

	ownerSql = `SELECT tr.ReceiverId, tr.Time FROM Transactions AS tr
				  WHERE tr.ResourceId = ? AND tr.SimId = ?
				  ORDER BY tr.Time ASC;`
	rootsSql = `SELECT res.ResourceId,res.TimeCreated,rc.AgentId,res.QualId,Quantity FROM Resources AS res
				  INNER JOIN ResCreators AS rc ON res.ResourceId = rc.ResourceId
				  WHERE res.SimId = ? AND rc.SimId = ?;`
)

// Prepare creates necessary indexes and tables required for efficient
// calculation of cyclus simulation inventory information.  Should be called
// once before walking begins.
func Prepare(conn *sqlite3.Conn) (err error) {
	for _, sql := range preExecStmts {
		if err := conn.Exec(sql); err != nil {
			log.Println("    ", err)
		}
	}
	return nil
}

// Finish should be called for a cyclus database after all walkers have
// completed processing inventory data. It creates final indexes and other
// finishing tasks.
func Finish(conn *sqlite3.Conn) (err error) {
	for _, sql := range postExecStmts {
		if err := conn.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

type Node struct {
	ResId     int
	OwnerId   int
	StartTime int
	EndTime   int
	QualId    int
	Quantity  float64
}

// Context encapsulates the logic for building a fast, queryable inventories
// table for a specific simulation from raw cyclus output database.
type Context struct {
	*sqlite3.Conn
	// Simid is the cyclus simulation id targeted by this context.  Must be
	// set.
	Simid       []byte
	Log         *log.Logger
	mappednodes map[int32]struct{}
	tmpResTbl   string
	tmpResStmt  *sqlite3.Stmt
	dumpStmt    *sqlite3.Stmt
	ownerStmt   *sqlite3.Stmt
	resCount    int
	nodes       []*Node
}

func NewContext(conn *sqlite3.Conn, simid []byte, history chan string) *Context {
	return &Context{
		Conn:  conn,
		Simid: simid,
		Log:   log.New(NullWriter{}, "", 0),
	}
}

func (c *Context) init() {
	// skip if the post processing already exists for this simid in the db
	_, err := c.Query("SELECT * FROM Agents WHERE SimId = ? LIMIT 1", c.Simid)
	if err == nil {
		panic(fmt.Sprintf("SimId %x is already post-processed. Skipping.\n", c.Simid))
	} else if err != io.EOF {
		panicif(err)
	}

	err = c.Exec("BEGIN TRANSACTION;")
	panicif(err)
	defer c.Exec("END TRANSACTION;")

	// build Agents table
	sql := `INSERT INTO Agents
				SELECT n.SimId,n.AgentId,n.Kind,n.Spec,n.Prototype,n.ParentId,n.Lifetime,n.EnterTime,x.ExitTime
				FROM
					AgentEntry AS n
					LEFT JOIN AgentExit AS x ON n.AgentId = x.AgentId AND n.SimId = x.SimId AND n.SimId = ?;`
	err = c.Exec(sql, c.Simid)
	panicif(err)

	c.nodes = make([]*Node, 0, 10000)
	c.mappednodes = map[int32]struct{}{}

	// build TimeList table
	sql = "SELECT Duration FROM Info WHERE SimId = ?;"
	for stmt, err := c.Query(sql, c.Simid); err == nil; err = stmt.Next() {
		var dur int
		panicif(stmt.Scan(&dur))
		for i := 0; i < dur; i++ {
			c.Exec("INSERT INTO TimeList VALUES (?, ?);", c.Simid, i)
		}
	}
	if err != io.EOF {
		panicif(err)
	}

	// create temp res table without simid
	c.Log.Println("Creating temporary resource table...")
	c.tmpResTbl = "tmp_restbl_" + fmt.Sprintf("%x", c.Simid)
	err = c.Exec("DROP TABLE IF EXISTS " + c.tmpResTbl)
	panicif(err)

	sql = "CREATE TABLE " + c.tmpResTbl + " AS SELECT ResourceId,TimeCreated,Parent1,Parent2,QualId,Quantity FROM Resources WHERE SimId = ?;"
	err = c.Exec(sql, c.Simid)
	panicif(err)

	c.Log.Println("Indexing temporary resource table...")
	err = c.Exec(query.Index(c.tmpResTbl, "Parent1"))
	panicif(err)

	err = c.Exec(query.Index(c.tmpResTbl, "Parent2"))
	panicif(err)

	// create prepared statements
	c.tmpResStmt, err = c.Prepare(resSqlHead + c.tmpResTbl + resSqlTail)
	panicif(err)

	c.dumpStmt, err = c.Prepare(dumpSql)
	panicif(err)

	c.ownerStmt, err = c.Prepare(ownerSql)
	panicif(err)
}

// WalkAll constructs the inventories table in the cyclus database alongside
// other tables. Creates several indexes in the process.  Finish should be
// called on the database connection after all simulation id's have been
// walked.
func (c *Context) WalkAll() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	c.Log.Printf("--- Building inventories for simid %x ---\n", c.Simid)
	c.init()

	c.Log.Println("Retrieving root resource nodes...")
	roots := c.getRoots()

	c.Log.Printf("Found %v root nodes\n", len(roots))
	for i, n := range roots {
		c.Log.Printf("    Processing root %d...\n", i)
		c.walkDown(n)
	}

	c.Log.Println("Dropping temporary resource table...")
	err = c.Exec("DROP TABLE " + c.tmpResTbl)
	panicif(err)

	c.dumpNodes()

	return nil
}

func (c *Context) getRoots() (roots []*Node) {
	sql := "SELECT COUNT(*) FROM ResCreators WHERE SimId = ?"
	stmt, err := c.Query(sql, c.Simid)
	panicif(err)

	n := 0
	err = stmt.Scan(&n)
	panicif(err)
	stmt.Reset()

	roots = make([]*Node, 0, n)
	for stmt, err = c.Query(rootsSql, c.Simid, c.Simid); err == nil; err = stmt.Next() {
		node := &Node{EndTime: math.MaxInt32}
		err := stmt.Scan(&node.ResId, &node.StartTime, &node.OwnerId, &node.QualId, &node.Quantity)
		panicif(err)

		roots = append(roots, node)
	}
	if err != io.EOF {
		panic(err)
	}
	return roots
}

func (c *Context) walkDown(node *Node) {
	if _, ok := c.mappednodes[int32(node.ResId)]; ok {
		return
	}
	c.mappednodes[int32(node.ResId)] = struct{}{}

	// dump if necessary
	c.resCount++
	if c.resCount%DumpFreq == 0 {
		c.dumpNodes()
	}

	// find resource's children
	kids := make([]*Node, 0, 2)
	err := c.tmpResStmt.Query(node.ResId, node.ResId)
	for ; err == nil; err = c.tmpResStmt.Next() {
		child := &Node{EndTime: math.MaxInt32}
		err := c.tmpResStmt.Scan(&child.ResId, &child.StartTime, &child.QualId, &child.Quantity)
		panicif(err)
		node.EndTime = child.StartTime
		kids = append(kids, child)
	}
	if err != io.EOF {
		panic(err)
	}

	// find resources owner changes (that occurred before children)
	owners, times := c.getNewOwners(node.ResId)

	childOwner := node.OwnerId
	if len(owners) > 0 {
		node.EndTime = times[0]
		childOwner = owners[len(owners)-1]

		lastend := math.MaxInt32
		if len(kids) > 0 {
			lastend = kids[0].StartTime
		}
		times = append(times, lastend)
		for i := range owners {
			n := &Node{ResId: node.ResId,
				OwnerId:   owners[i],
				StartTime: times[i],
				EndTime:   times[i+1],
				QualId:    node.QualId,
				Quantity:  node.Quantity,
			}
			c.nodes = append(c.nodes, n)
		}
	}

	c.nodes = append(c.nodes, node)

	// walk down resource's children
	for _, child := range kids {
		child.OwnerId = childOwner
		c.walkDown(child)
	}
}

func (c *Context) getNewOwners(id int) (owners, times []int) {
	var owner, t int
	err := c.ownerStmt.Query(id, c.Simid)
	for ; err == nil; err = c.ownerStmt.Next() {
		err := c.ownerStmt.Scan(&owner, &t)
		panicif(err)

		if id == owner {
			continue
		}
		owners = append(owners, owner)
		times = append(times, t)
	}
	if err != io.EOF {
		panic(err)
	}
	return owners, times
}

func (c *Context) dumpNodes() {
	c.Log.Printf("    Dumping inventories (%d resources done)...\n", c.resCount)
	err := c.Exec("BEGIN TRANSACTION;")
	panicif(err)

	for _, n := range c.nodes {
		if n.EndTime > n.StartTime {
			err = c.dumpStmt.Exec(c.Simid, n.ResId, n.OwnerId, n.StartTime, n.EndTime, n.QualId, n.Quantity)
			panicif(err)
		}
	}
	err = c.Exec("END TRANSACTION;")
	panicif(err)

	c.nodes = c.nodes[:0]
}
