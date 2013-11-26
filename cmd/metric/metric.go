package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strconv"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/rwcarlsen/cyan/query"
)

var (
	help   = flag.Bool("h", false, "Print this help message")
	dbname = flag.String("db", "", "cyclus sqlite database to query")
	simid  = flag.String("simid", "", "simulation id (empty string defaults to first sim id in database")
)

var command string

var db *sql.DB

func main() {
	log.SetFlags(0)
	flag.Parse()

	// global flags
	if *help || flag.NArg() < 1 {
		fmt.Println("Usage: metric -db <cyclus-db> [opts] <command> [args...]")
		fmt.Println("Calculates metrics for cyclus simulation data in a sqlite database.")
		flag.PrintDefaults()
		return
	} else if *dbname == "" {
		log.Fatal("must specify database with db flag")
	}

	var err error
	db, err = sql.Open("sqlite3", *dbname)
	fatalif(err)
	defer db.Close()

	if *simid == "" {
		ids, err := query.SimIds(db)
		fatalif(err)
		*simid = ids[0]
	}

	switch flag.Arg(0) {
	case "sims":
		ids, err := query.SimIds(db)
		fatalif(err)
		for _, id := range ids {
			info, err := query.SimStat(db, id)
			fatalif(err)
			fmt.Println(info)
		}
	case "agents":
		ags, err := query.AllAgents(db, *simid)
		fatalif(err)
		for _, a := range ags {
			fmt.Println(a)
		}
	case "inv":
		doInv()
	case "created":
		doCreated()
	case "flow":
		doFlow()
	case "test":
		test()
	default:
		log.Fatalf("unrecognized command %v", flag.Arg(0))
	}
}

func test() {
	fmt.Println("running test func...")
	sql := `SELECT ti.Time,cmp.IsoID,SUM(cmp.Quantity * res.Quantity)
			   FROM Resources AS res
			   INNER JOIN Inventories AS inv ON inv.ResID = res.ID
			   INNER JOIN Compositions AS cmp ON cmp.ID = res.StateID
			   INNER JOIN Agents AS ag ON ag.ID = inv.AgentID
			   INNER JOIN Times AS ti ON (ti.Time >= inv.StartTime AND ti.Time < inv.EndTime)
			WHERE
				ag.Prototype = 'LWR_Reactor'
				AND cmp.IsoID = 92235
				AND res.SimID = '653de5fc-7422-41cd-b22b-557244e0756b'
				AND cmp.SimID = res.SimID AND ag.SimID = res.SimID AND inv.SimID = res.SimID
			GROUP BY ti.Time,cmp.IsoID;`
	rows, err := db.Query(sql)
	fatalif(err)
	var t, iso int
	var qty float64
	for rows.Next() {
		rows.Scan(&t, &iso, &qty)
		fmt.Printf("%v\t%v\t%v\n", t, iso, qty)
	}
	fatalif(rows.Err())
}

func doInv() {
	var err error
	t := -1
	var agents []int

	switch n := flag.NArg(); {
	case n == 2:
		t, err = strconv.Atoi(flag.Arg(1))
		fatalif(err)
		fallthrough
	case n > 2:
		for _, arg := range flag.Args()[2:] {
			id, err := strconv.Atoi(arg)
			fatalif(err)
			agents = append(agents, id)
		}
	}

	m, err := query.InvAt(db, *simid, t, agents...)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

func doCreated() {
	var err error
	t0, t1 := -1, -1
	var agents []int

	switch n := flag.NArg(); {
	case n == 2:
		t0, err = strconv.Atoi(flag.Arg(1))
		fatalif(err)
		fallthrough
	case n == 3:
		t1, err = strconv.Atoi(flag.Arg(2))
		fatalif(err)
		fallthrough
	case n > 3:
		for _, arg := range flag.Args()[3:] {
			id, err := strconv.Atoi(arg)
			fatalif(err)
			agents = append(agents, id)
		}
	}

	m, err := query.MatCreated(db, *simid, t0, t1, agents...)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

func doFlow() {
	var err error
	t0, t1 := -1, -1
	var from []int
	var to []int

	if flag.NArg() < 6 {
		log.Fatal("need 4 args for flow: <t0> <t1> <fromAgents...> .. <toAgents...>")
	}

	t0, err = strconv.Atoi(flag.Arg(1))
	fatalif(err)
	t1, err = strconv.Atoi(flag.Arg(2))
	fatalif(err)

	before := true
	for _, arg := range flag.Args()[3:] {
		if arg == ".." {
			before = false
			continue
		}

		id, err := strconv.Atoi(arg)
		fatalif(err)
		if before {
			from = append(from, id)
		} else {
			to = append(to, id)
		}
	}
	if len(to) < 1 {
		log.Fatal("no receiving agents specified. Use <fromAgents> .. <toAgents>")
	}

	m, err := query.Flow(db, *simid, t0, t1, from, to)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
