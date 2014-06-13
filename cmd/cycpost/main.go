package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mxk/go-sqlite/sqlite3"
	"github.com/rwcarlsen/cyan/post"
)

var help = flag.Bool("h", false, "Print this help message.")
var verbose = flag.Bool("v", false, "print verbose progress output")

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *help || flag.NArg() != 1 {
		fmt.Println("Usage: inventory [cyclus-db]")
		fmt.Println("Creates a fast queryable inventory table for a cyclus sqlite output file.\n")
		flag.PrintDefaults()
		return
	}

	fname := flag.Arg(0)

	db, err := sql.Open("sqlite3", fname)
	fatalif(err)
	defer db.Close()

	fatalif(post.Prepare(db))
	defer post.Finish(db)

	simids, err := post.GetSimIds(db)
	fatalif(err)

	for _, simid := range simids {
		ctx := post.NewContext(db, simid)
		if *verbose {
			ctx.Log = log.New(os.Stdout, "", 0)
		}
		err := ctx.WalkAll()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
