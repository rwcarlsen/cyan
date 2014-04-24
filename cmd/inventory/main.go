package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mxk/go-sqlite/sqlite3"
	"github.com/rwcarlsen/cyan/post"
)

var help = flag.Bool("h", false, "Print this help message.")

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

	conn, err := sqlite3.Open(fname)
	fatalif(err)
	defer conn.Close()

	fatalif(post.Prepare(conn))
	defer post.Finish(conn)

	simids, err := post.GetSimIds(conn)
	fatalif(err)

	for _, simid := range simids {
		ctx := post.NewContext(conn, simid, nil)
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
