package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/rwcarlsen/cyan/gone"
)

var shownuc = flag.Bool("shownuc", true, "print corresponding nuclide next to each value")
var decay = flag.Bool("decay", false, "print decay constants")
var energy = flag.String("energy", "thermal", "thermal, fission")
var yield = flag.String("yield", "", "print fission yields from `from-nuc` to given nuclides")
var list = flag.Bool("list", false, "print a list of all nuclides")
var nonzero = flag.Bool("nonzero", false, "only show nonzero values if shownuc is true")

func main() {
	flag.Parse()

	nucstrs := flag.Args()
	if len(nucstrs) == 0 && !*list {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		nucstr := strings.Split(string(data), "\n")
		for _, nuc := range nucstr {
			if len(nuc) > 0 {
				nucstrs = append(nucstrs, nuc)
			}
		}
	}

	nucs := []gone.Nuc{}
	for _, nuc := range nucstrs {
		nucs = append(nucs, gone.Id(nuc))
	}

	switch {
	case *list:
		for i := 1001; i < 150000; i++ {
			nuc := gone.IdFromInt(i * 10000)
			if nuc > 0 {
				fmt.Println(nuc)
			}
			nuc = gone.IdFromInt(i*10000 + 1)
			if nuc > 0 {
				fmt.Println(nuc)
			}
			nuc = gone.IdFromInt(i*10000 + 2)
			if nuc > 0 {
				fmt.Println(nuc)
			}
			nuc = gone.IdFromInt(i*10000 + 3)
			if nuc > 0 {
				fmt.Println(nuc)
			}
		}
	case *decay:
		for _, nuc := range nucs {
			v := gone.DecayConst(nuc)
			if *shownuc {
				if !*nonzero || v > 0 {
					fmt.Printf("%v %v\n", nuc, v)
				}
			} else {
				fmt.Printf("%v\n", v)
			}
		}
	case len(*yield) > 0:
		from := gone.Id(*yield)
		e := gone.Thermal
		if *energy == "fission" {
			e = gone.FastFission
		}
		for _, nuc := range nucs {
			v := gone.FissProdYield(from, nuc, e)
			if *shownuc {
				if !*nonzero || v > 0 {
					fmt.Printf("%v %v\n", nuc, v)
				}
			} else {
				fmt.Printf("%v\n", v)
			}
		}
	default:
		for _, nuc := range nucs {
			fmt.Println(nuc)
		}
	}
}
