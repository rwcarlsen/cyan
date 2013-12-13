package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sort"
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

var cmds = NewCmdSet()

func init() {
	cmds.Register("agents", "list all agents in the simulation", doAgents)
	cmds.Register("sims", "list all simulations in the database", doSims)
	cmds.Register("inv", "show inventory of one or more agents at a specific timestep", doInv)
	cmds.Register("created", "show material created by one or more agents between specific timesteps", doCreated)
	cmds.Register("deployseries", "print a time-series of a prototype's total active deployments", doDeploySeries)
	cmds.Register("flow", "Show total transacted material between two groups of agents between specific timesteps", doFlow)
	cmds.Register("invseries", "print a time series of an agent's inventory for specified isotopes", doInvSeries)
	cmds.Register("flowgraph", "print a graphviz dot graph of resource arcs between facilities", doFlowGraph)
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *help || flag.NArg() < 1 {
		fmt.Println("Usage: metric -db <cyclus-db> [opts] <command> [args...]")
		fmt.Println("Calculates metrics for cyclus simulation data in a sqlite database.")
		flag.PrintDefaults()
		fmt.Println("\nCommands:")
		for i := range cmds.Names {
			fmt.Printf("    - %v: %v\n", cmds.Names[i], cmds.Helps[i])
		}
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

	cmds.Execute(flag.Args())
}

func doSims(args []string) {
	ids, err := query.SimIds(db)
	fatalif(err)
	for _, id := range ids {
		info, err := query.SimStat(db, id)
		fatalif(err)
		fmt.Println(info)
	}
}

func doAgents(args []string) {
	ags, err := query.AllAgents(db, *simid)
	fatalif(err)
	for _, a := range ags {
		fmt.Println(a)
	}
}

func doInv(args []string) {
	fs := flag.NewFlagSet("inv", flag.ExitOnError)
	t := fs.Int("t", -1, "timestep of inventory (-1 = end of simulation)")
	fs.Usage = func() {
		log.Print("Usage: inv [agent-id...]\nZero agents uses all agent inventories")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	var agents []int
	for _, arg := range fs.Args() {
		id, err := strconv.Atoi(arg)
		fatalif(err)
		agents = append(agents, id)
	}

	m, err := query.InvAt(db, *simid, *t, agents...)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

type Row struct {
	X  int
	Ys []float64
}

type MultiSeries [][]query.XY

func (ms MultiSeries) Rows() []Row {
	rowmap := map[int]Row{}
	xs := []int{}
	for i, s := range ms {
		for _, xy := range s {
			row, ok := rowmap[xy.X]
			if !ok {
				xs = append(xs, xy.X)
				row.Ys = make([]float64, len(ms))
				row.X = xy.X
			}
			row.Ys[i] = xy.Y
			rowmap[xy.X] = row
		}
	}

	sort.Ints(xs)
	rows := make([]Row, 0, len(rowmap))
	for _, x := range xs {
		rows = append(rows, rowmap[x])
	}
	return rows
}

func doInvSeries(args []string) {
	fs := flag.NewFlagSet("invseries", flag.ExitOnError)
	fs.Usage = func() { log.Print("Usage: invseries <agent-id> <isotope> [isotope...]"); fs.PrintDefaults() }
	fs.Parse(args)
	if fs.NArg() < 2 {
		fs.Usage()
		return
	}

	agent, err := strconv.Atoi(fs.Arg(0))
	fatalif(err)

	isos := []int{}
	for _, arg := range fs.Args()[1:] {
		iso, err := strconv.Atoi(arg)
		fatalif(err)
		isos = append(isos, iso)
	}

	ms := MultiSeries{}
	for _, iso := range isos {
		xys, err := query.InvSeries(db, *simid, agent, iso)
		ms = append(ms, xys)
		fatalif(err)
	}

	fmt.Printf("# Agent %v inventory in kg\n", agent)
	fmt.Printf("# [Timestep]")
	for _, iso := range isos {
		fmt.Printf(" [%v]", iso)
	}
	fmt.Printf("\n")
	for _, row := range ms.Rows() {
		fmt.Printf("%v", row.X)
		for _, y := range row.Ys {
			fmt.Printf(" %v ", y)
		}
		fmt.Printf("\n")
	}
}

func doFlowGraph(args []string) {
	fs := flag.NewFlagSet("flowgraph", flag.ExitOnError)
	fs.Usage = func() { log.Print("Usage: flowgraph"); fs.PrintDefaults() }
	proto := fs.Bool("proto", false, "aggregate nodes by prototype")
	t0 := fs.Int("t1", -1, "beginning of time interval (default is beginning of simulation)")
	t1 := fs.Int("t2", -1, "end of time interval (default if end of simulation)")
	fs.Parse(args)

	arcs, err := query.FlowGraph(db, *simid, *t0, *t1, *proto)
	fatalif(err)

	fmt.Println("digraph ResourceFlows {")
	fmt.Println("    overlap = false;")
	fmt.Println("    nodesep=1.0;")
	fmt.Println("    edge [fontsize=9];")
	for _, arc := range arcs {
		fmt.Printf("    \"%v\" -> \"%v\" [label=\"%v\\n(%.3g kg)\"];\n", arc.Src, arc.Dst, arc.Commod, arc.Quantity)
	}
	fmt.Println("}")
}

func doDeploySeries(args []string) {
	fs := flag.NewFlagSet("deployseries", flag.ExitOnError)
	fs.Usage = func() { log.Print("Usage: deployseries <prototype>"); fs.PrintDefaults() }
	fs.Parse(args)
	if fs.NArg() < 1 {
		fs.Usage()
		return
	}

	proto := fs.Arg(0)
	xys, err := query.DeployCumulative(db, *simid, proto)
	fatalif(err)

	fmt.Printf("# Prototype %v total active deployments\n", proto)
	fmt.Println("# [Timestep] [Count]")
	for _, xy := range xys {
		fmt.Printf("%v %v\n", xy.X, xy.Y)
	}
}

func doCreated(args []string) {
	fs := flag.NewFlagSet("created", flag.ExitOnError)
	fs.Usage = func() { log.Print("Usage: created [agent-id...]\nZero agents uses all agents"); fs.PrintDefaults() }
	t0 := fs.Int("t1", -1, "beginning of time interval (default is beginning of simulation)")
	t1 := fs.Int("t2", -1, "end of time interval (default if end of simulation)")
	fs.Parse(args)

	var agents []int

	for _, arg := range fs.Args() {
		id, err := strconv.Atoi(arg)
		fatalif(err)
		agents = append(agents, id)
	}

	m, err := query.MatCreated(db, *simid, *t0, *t1, agents...)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

func doFlow(args []string) {
	fs := flag.NewFlagSet("created", flag.ExitOnError)
	t0 := fs.Int("t1", -1, "beginning of time interval (default is beginning of simulation)")
	t1 := fs.Int("t2", -1, "end of time interval (default if end of simulation)")
	fs.Usage = func() {
		log.Print("Usage: flow <from-agents...> .. <to-agents...>\nZero agents uses all agents")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	var from []int
	var to []int

	if flag.NArg() < 3 {
		fs.Usage()
		return
	}

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
		fs.Usage()
		return
	}

	m, err := query.Flow(db, *simid, *t0, *t1, from, to)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type CmdSet struct {
	funcs map[string]func([]string)
	Names []string
	Helps []string
}

func NewCmdSet() *CmdSet {
	return &CmdSet{funcs: map[string]func([]string){}}
}

func (cs *CmdSet) Register(name, brief string, f func([]string)) {
	cs.Names = append(cs.Names, name)
	cs.Helps = append(cs.Helps, brief)
	cs.funcs[name] = f
}

func (cs *CmdSet) Execute(args []string) {
	cmd := args[0]
	f, ok := cs.funcs[cmd]
	if !ok {
		log.Fatalf("Invalid command '%v'", cmd)
	}
	f(args[1:])
}
