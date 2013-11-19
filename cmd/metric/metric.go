package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"github.com/rwcarlsen/cyan/query"
)

var (
	pid   = flag.Bool("pid", false, "Print simulation ids")
	help  = flag.Bool("h", false, "Print this help message")
	simid = flag.String("simid", "", "simulation id (empty string defaults to first sim id in database")

	invat     = flag.Int("inv", -1, "Print global inventory at specified time (-1 prints inventory at simulation end)")
	createinv = flag.Bool("created", false, "Print total of all created material")
	fpeat     = flag.Int("fpe", 0, "Print fission potential energy in Joules at specified simulation timestep")
)

var db *sql.DB

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *help || flag.NArg() < 1 {
		fmt.Println("Usage: metric [opts] <cyclus-db>")
		fmt.Println("Calculates global metrics for cyclus simulation data in an output database.")
		flag.PrintDefaults()
		return
	}

	fname := flag.Arg(0)

	var err error
	db, err = sql.Open("sqlite3", fname)
	fatalif(err)
	defer db.Close()

	if *pid {
		ids, err := query.SimIds(db)
		fatalif(err)
		for _, id := range ids {
			fmt.Println(id)
		}
		return
	}

	if *simid == "" {
		ids, err := query.SimIds(db)
		fatalif(err)
		*simid = ids[0]
	}

	flag.Visit(processFlags)

}

func processFlags(fg *flag.Flag) {
	switch fg.Name {
	case "inv":
		m, err := query.AllMatAt(db, *simid, *invat)
		fatalif(err)
		fmt.Printf("%+v\n", m)
	case "createinv":
		if *createinv {
			m, err := query.AllMatCreated(db, *simid)
			fatalif(err)
			fmt.Printf("%+v\n", m)
		}
	case "fpe":
		m, err := query.AllMatAt(db, *simid, *fpeat)
		fatalif(err)
		fmt.Printf("%v\n", m.FPE())
	default:
		log.Fatal("unrecognized flag ", fg.Name)
	}
}

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
