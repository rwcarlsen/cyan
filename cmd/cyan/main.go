package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"

	"code.google.com/p/go-uuid/uuid"
	"github.com/rwcarlsen/cyan/nuc"
	"github.com/rwcarlsen/cyan/post"
	"github.com/rwcarlsen/cyan/query"
	_ "github.com/rwcarlsen/go-sqlite3"
)

var (
	custom    = flag.String("custom", "", "path to custom sql query spec file")
	showquery = flag.Bool("query", false, "show query SQL for a subcommand instead of executing it")
	dbname    = flag.String("db", "", "cyclus sqlite database to query")
	simidstr  = flag.String("simid", "", "simulation id in hex (empty string defaults to first sim id in database")
	noheader  = flag.Bool("noheader", false, "don't print header line with output data")
)

var simid []byte

var command string

var db *sql.DB

var cmds = NewCmdSet()

// map[cmdname]sqltext
var customSql = map[string]string{}

func plot(data *bytes.Buffer, style string, xlabel, ylabel, title string) {
	s := ""
	s += `set xlabel '{{.Xlabel}}';`
	s += `set ylabel '{{.Ylabel}}';`
	s += `plot '-' every ::2 using 1:2 with {{.Style}} title '{{.Title}}';`
	s += `pause -1`

	tmpl := template.Must(template.New("gnuplot").Parse(s))
	var buf bytes.Buffer
	config := struct{ Style, Xlabel, Ylabel, Title string }{style, xlabel, ylabel, title}
	err := tmpl.Execute(&buf, config)
	fatalif(err)

	cmd := exec.Command("gnuplot", "-p", "-e", buf.String())
	cmd.Stdin = data
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fatalif(cmd.Run())
}

func init() {
	cmds.RegisterDiv("General")
	cmds.Register("sims", "list all simulations in the database", doSims)
	cmds.Register("infile", "show the simulation's input file", doInfile)
	cmds.Register("version", "show simulation's cyclus version info", doVersion)
	cmds.Register("post", "post process the database", doPost)
	cmds.Register("table", "show the contents of a specific table", doTable)
	cmds.Register("ts", "investigate time-series data tables", doTimeSeries)
	cmds.RegisterDiv("Agents")
	cmds.Register("agents", "list all agents in the simulation", doAgents)
	cmds.Register("protos", "list all prototypes in the simulation", doProtos)
	cmds.Register("deployed", "time series total active deployments by prototype", doDeployed)
	cmds.Register("built", "time series of new builds by prototype", doBuilt)
	cmds.Register("decom", "time series of a decommissionings by prototype", doDecom)
	cmds.Register("ages", "list ages of agents at a particular time step", doAges)
	cmds.RegisterDiv("Flow")
	cmds.Register("commods", "show commodity transaction counts and quantities", doCommods)
	cmds.Register("flow", "time series of material transacted between agents", doFlow)
	cmds.Register("flowgraph", "generate a graphviz dot script of flows between agents", doFlowGraph)
	cmds.Register("trans", "time series of transaction quantity over time", doTrans)
	cmds.RegisterDiv("Other")
	cmds.Register("inv", "time series of inventory by prototype", doInv)
	cmds.Register("power", "time series of power produced", doPower)
	cmds.Register("energy", "thermal energy (J) generated between 2 timesteps", doEnergy)
	cmds.Register("created", "material created by agents between 2 timesteps", doCreated)
}

func initdb() {
	if *showquery {
		// don't need a database for printing queries
		return
	} else if *dbname == "" {
		log.Fatal("must specify database with -db flag")
	}

	var err error
	db, err = sql.Open("sqlite3", *dbname)
	fatalif(err)

	if *simidstr == "" {
		ids, err := query.SimIds(db)
		fatalif(err)
		simid = ids[0]
	} else {
		simid = uuid.Parse(*simidstr)
		if simid == nil {
			log.Fatalf("invalid simid '%s'", *simidstr)
		}
	}

	post.Process(db)
}

func main() {
	log.SetFlags(0)
	flag.CommandLine.Usage = func() {
		fmt.Println("Usage: cyan [-db <cyclus-db>] [flags...] <subcommand> [flags...] [args...]")
		fmt.Println("Computes metrics for cyclus simulation data in a sqlite database.")
		fmt.Println("\nOptions:")
		flag.CommandLine.PrintDefaults()
		fmt.Println("\nSub-commands:")
		tw := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
		for i := range cmds.Names {
			if cmds.IsDiv(i) {
				fmt.Fprintf(tw, "\n\t[%v]\n", cmds.Names[i])
			} else {
				fmt.Fprintf(tw, "\t\t%v\t%v\n", cmds.Names[i], cmds.Helps[i])
			}
		}
		tw.Flush()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: cyan -db <cyclus-db> [flags...] <command> [flags...] [args...]")
		fmt.Println("Computes metrics for cyclus simulation data in a sqlite database.")
		flag.PrintDefaults()
		fmt.Println("\nSub-commands:")
		for i := range cmds.Names {
			fmt.Printf("    %v: %v\n", cmds.Names[i], cmds.Helps[i])
		}
		return
	}

	if *custom != "" {
		data, err := ioutil.ReadFile(*custom)
		fatalif(err)
		fatalif(json.Unmarshal(data, &customSql))
	}

	// run command
	cmds.Execute(flag.Args())
}

func doCustom(w io.Writer, cmd string, args ...interface{}) {
	s, ok := customSql[cmd]
	if !ok {
		log.Fatalf("Invalid command/query %v", cmd)
	} else if *showquery {
		fmt.Fprint(w, s)
		return
	}

	rows, err := db.Query(s, args...)
	fatalif(err)

	tw := tabwriter.NewWriter(w, 4, 4, 1, ' ', 0)
	cols, err := rows.Columns()
	fatalif(err)

	simidcol := -1
	if !*noheader {
		// write header line
		for i, c := range cols {
			if strings.Contains(strings.ToLower(c), "simid") {
				simidcol = i
			}
			_, err := tw.Write([]byte(c + "\t"))
			fatalif(err)
		}
		_, err = tw.Write([]byte("\n"))
		fatalif(err)
	}

	vs := make([]interface{}, len(cols))
	vals := make([]*sql.NullString, len(cols))
	for i := range vals {
		vals[i] = &sql.NullString{}
		vs[i] = vals[i]
	}

	for rows.Next() {
		for i := range vals {
			vals[i].Valid = false
		}

		err := rows.Scan(vs...)
		fatalif(err)

		for i, v := range vals {
			if v.Valid {
				s = v.String
				if i == simidcol {
					s = uuid.UUID(v.String).String()
				}
				tw.Write([]byte(s + "\t"))
			} else {
				tw.Write([]byte("NULL\t"))
			}
		}

		_, err = tw.Write([]byte("\n"))
		fatalif(err)
	}
	fatalif(rows.Err())
	fatalif(tw.Flush())
}

func doSims(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()
	s := `
SELECT i.SimId AS SimId,Duration,Handle,Decay
FROM Info As i
JOIN DecayMode AS d ON i.SimId=d.SimId 
`
	customSql[cmd] = s
	doCustom(os.Stdout, cmd)
}

func doVersion(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	av := fs.Bool("agent", false, "show simulation's agent versions instead")
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()
	s := `
SELECT
i.CyclusVersionDescribe AS Cyclus,
i.SqliteVersion As SQLite,
i.Hdf5Version AS HDF5,
i.BoostVersion AS Boost,
i.LibXML2Version AS libxml2,
i.CoinCBCVersion AS CoinCBC,
x.LibXMLPlusPlusVersion AS 'libxml++'
FROM info AS i
JOIN xmlppinfo AS x ON i.simid=x.simid
WHERE i.simid=?
`
	if *av {
		s = `SELECT Spec AS Archetype,Version FROM AgentVersions WHERE simid=?`
	}
	customSql[cmd] = s
	doCustom(os.Stdout, cmd, simid)
}

func doPost(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()
}

func doInfile(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()
	s := `
SELECT data FROM inputfiles WHERE simid=?;
`
	var buf bytes.Buffer
	customSql[cmd] = s
	doCustom(&buf, cmd, simid)
	data := buf.String()
	data = strings.Replace(data, "Data", "", 1)
	data = strings.TrimLeft(data, "\r\n\t ")
	fmt.Print(data)
}

func doAgents(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	proto := fs.String("proto", "", "filter by prototype (default is all prototypes)")
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	iargs := []interface{}{simid}
	s := `
SELECT AgentId,Kind,Prototype,ParentId,EnterTime,ExitTime,Lifetime
FROM Agents
WHERE SimId = ?
`
	if *proto != "" {
		s += `AND Prototype = ?`
		iargs = append(iargs, *proto)
	}
	customSql[cmd] = s
	doCustom(os.Stdout, cmd, iargs...)
}

func doAges(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	proto := fs.String("proto", "", "filter by prototype (default is all prototypes)")
	fs.Usage = func() {
		log.Printf("Usage: %v [time-step]", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	t, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		log.Fatalf("invalid time step '%v')", fs.Arg(0))
	}

	iargs := []interface{}{t, simid, t, t}
	s := `
SELECT ? - a.entertime AS Age FROM Agents as a
WHERE a.simid=?
AND a.entertime <= ?
AND (a.exittime >= ? OR a.exittime ISNULL)
`
	if *proto != "" {
		s += `AND Prototype = ?`
		iargs = append(iargs, *proto)
	}

	customSql[cmd] = s
	doCustom(os.Stdout, cmd, iargs...)
}

func doTable(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v [table-name]", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		log.Printf("If no table name is given, prints a list of tables.")
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	if fs.NArg() > 0 {
		s := "SELECT * FROM " + fs.Arg(0) + " WHERE SimId = ?"
		customSql[cmd] = s
		doCustom(os.Stdout, cmd, simid)
	} else {
		s := "SELECT name FROM sqlite_master WHERE type='table';"
		customSql[cmd] = s
		doCustom(os.Stdout, cmd)
	}
}

func doTimeSeries(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	proto := fs.String("proto", "", "filter by prototype (default is all prototypes)")
	plotit := fs.Bool("p", false, "plot the data")
	fs.Usage = func() {
		log.Printf("Usage: %v [table-name]", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		log.Printf("If no series name is given, prints a list of all time series'.")
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	if fs.NArg() == 0 {
		s := "SELECT name as foobarbaz FROM sqlite_master WHERE type='table' AND instr(name,'TimeSeries');"
		customSql[cmd] = s
		var buf bytes.Buffer
		doCustom(&buf, cmd)
		data := buf.String()
		data = strings.Replace(data, "foobarbaz", "", -1)
		data = strings.Replace(data, "TimeSeries", "", -1)
		data = strings.TrimLeft(data, "\r\n\t ")
		fmt.Print(string(data))
	} else {
		tsname := fs.Arg(0)
		filter := ""
		s := `
SELECT tl.Time AS Time,IFNULL(sub.Val,0) AS {{.Name}}
FROM timelist as tl LEFT JOIN (
	SELECT p.simid AS simid,p.Time AS Time,TOTAL(p.Value) AS Val
	FROM timeseries{{.Name}} AS p
	JOIN agents as a on a.agentid=p.agentid AND a.simid=p.simid
	WHERE p.simid=? {{.Filter}}
	GROUP BY p.Time
) AS sub ON tl.time=sub.time AND tl.simid=sub.simid
`

		tmpl := template.Must(template.New("sql").Parse(s))
		var buf bytes.Buffer
		if *proto != "" {
			filter = " AND a.prototype='" + *proto + "' "
		}
		tmpl.Execute(&buf, struct{ Name, Filter string }{tsname, filter})
		customSql[cmd] = buf.String()

		var buff bytes.Buffer
		doCustom(&buff, cmd, simid)
		if *plotit {
			plot(&buff, "linespoints", "Time (Months)", "Power (MWe)", "Total Power Produced")
		} else {
			fmt.Print(buff.String())
		}
	}
}

func doPower(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	proto := fs.String("proto", "", "filter by prototype (default is all prototypes)")
	plotit := fs.Bool("p", false, "plot the data")
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	s := `
SELECT tl.Time AS Time,IFNULL(sub.Power,0) AS Power
FROM timelist as tl LEFT JOIN (
	SELECT p.simid AS simid,p.Time AS Time,TOTAL(p.Value) AS Power
	FROM timeseriespower AS p
	JOIN agents as a on a.agentid=p.agentid AND a.simid=p.simid
	WHERE p.simid=? {{.}}
	GROUP BY p.Time
) AS sub ON tl.time=sub.time AND tl.simid=sub.simid
`

	tmpl := template.Must(template.New("sql").Parse(s))
	var buf bytes.Buffer
	if *proto == "" {
		tmpl.Execute(&buf, "")
	} else {
		tmpl.Execute(&buf, " AND a.prototype='"+*proto+"' ")
	}
	customSql[cmd] = buf.String()

	var buff bytes.Buffer
	doCustom(&buff, cmd, simid)
	if *plotit {
		plot(&buff, "linespoints", "Time (Months)", "Power (MWe)", "Total Power Produced")
	} else {
		fmt.Print(buff.String())
	}
}

func doDeployed(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v <prototype>", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	plotit := fs.Bool("p", false, "plot the data")
	fs.Parse(args)
	if fs.NArg() < 1 {
		log.Fatal("must specify a prototype")
	}
	initdb()

	proto := fs.Arg(0)
	s := `
SELECT tl.Time AS Time,IFNULL(n, 0) AS N_Deployed
FROM timelist AS tl
LEFT JOIN (
    SELECT tl.time AS time,COUNT(a.agentid) AS n
	FROM timelist AS tl
    LEFT JOIN agents AS a ON a.entertime <= tl.time AND (a.exittime >= tl.time OR a.exittime ISNULL) AND (tl.time < a.entertime + a.lifetime) AND a.simid=tl.simid
    WHERE a.simid=? AND a.prototype=?
    GROUP BY tl.time
) AS sub ON sub.time=tl.time
WHERE tl.simid=?
`
	customSql[cmd] = s
	var buf bytes.Buffer
	doCustom(&buf, cmd, simid, proto, simid)
	if *plotit {
		plot(&buf, "linespoints", "Time (Months)", "Number "+proto+" Deployed", "Deployed Facilities")
	} else {
		fmt.Print(buf.String())
	}
}

func doBuilt(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v <prototype>", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	plotit := fs.Bool("p", false, "plot the data")
	fs.Parse(args)
	if fs.NArg() < 1 {
		log.Fatal("must specify a prototype")
	}
	initdb()

	proto := fs.Arg(0)
	s := `
SELECT tl.time AS Time,ifnull(sub.n, 0) AS N_Built
FROM timelist AS tl
LEFT JOIN (
	SELECT a.simid,tl.time AS time,COUNT(a.agentid) AS n
	FROM agents AS a
	JOIN timelist AS tl ON tl.time=a.entertime AND tl.simid=a.simid
	WHERE a.simid=? AND a.prototype=?
	GROUP BY time
) AS sub ON tl.time=sub.time AND tl.simid=sub.simid
WHERE tl.simid=?
`

	customSql[cmd] = s
	var buf bytes.Buffer
	doCustom(&buf, cmd, simid, proto, simid)
	if *plotit {
		plot(&buf, "impulses", "Time (Months)", "Number "+proto+" Built", "New Facilities Built")
	} else {
		fmt.Print(buf.String())
	}
}

func doDecom(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v <prototype>", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	plotit := fs.Bool("p", false, "plot the data")
	fs.Parse(args)
	if fs.NArg() < 1 {
		log.Fatal("must specify a prototype")
	}
	initdb()

	proto := fs.Arg(0)
	s := `
SELECT tl.time AS Time,ifnull(sub.n, 0) AS N_Built
FROM timelist AS tl
LEFT JOIN (
	SELECT a.simid,tl.time AS time,COUNT(a.agentid) AS n
	FROM agents AS a
	JOIN timelist AS tl ON tl.time=a.exittime AND tl.simid=a.simid
	WHERE a.simid=? AND a.prototype=?
	GROUP BY time
) AS sub ON tl.time=sub.time AND tl.simid=sub.simid
WHERE tl.simid=?
`

	customSql[cmd] = s
	var buf bytes.Buffer
	doCustom(&buf, cmd, simid, proto, simid)
	if *plotit {
		plot(&buf, "impulses", "Time (Months)", "Number "+proto+" Decommissioned", "Facilities Decommissioned")
	} else {
		fmt.Print(buf.String())
	}
}

func doProtos(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	s := "SELECT DISTINCT Prototype FROM Prototypes WHERE simid=?;"
	customSql[cmd] = s
	doCustom(os.Stdout, cmd, simid)
}

func doCommods(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	s := `
SELECT Commodity,count(t.transactionid) AS N_Trans, TOTAL(r.quantity) AS Quantity
FROM transactions AS t
JOIN Resources AS r ON r.ResourceId=t.ResourceId AND r.SimId=t.SimId
WHERE r.simid=?
GROUP BY commodity;
`

	customSql[cmd] = s
	doCustom(os.Stdout, cmd, simid)
}

func doTrans(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	from := fs.String("from", "", "filter by supplying prototype")
	to := fs.String("to", "", "filter by receiving prototype")
	byagent := fs.Bool("byagent", false, "switch to/from filters to be agent IDs")
	nucs := fs.String("nucs", "", "filter by comma separated `nuclide`s")
	commod := fs.String("commod", "", "filter by a commodity")
	fs.Parse(args)
	initdb()

	s := `
SELECT t.time AS Time,t.SenderId AS SenderId,send.Prototype AS SenderProto,t.ReceiverId AS ReceiverId,recv.Prototype AS ReceiverProto,t.Commodity AS Commodity,SUM(r.Quantity*c.MassFrac) AS Quantity
FROM transactions AS t
JOIN resources AS r ON t.resourceid=r.resourceid AND r.simid=t.simid
JOIN agents AS send ON t.senderid=send.agentid AND send.simid=t.simid
JOIN agents AS recv ON t.receiverid=recv.agentid AND recv.simid=t.simid
JOIN compositions AS c ON c.qualid=r.qualid AND c.simid=t.simid
WHERE t.simid=? {{index . 0}} {{index . 1}} {{index . 2}} {{index . 3}}
GROUP BY t.transactionid
`

	filters := make([]string, 4)
	iargs := []interface{}{simid}
	if *from != "" {
		if *byagent {
			filters[0] = "AND t.senderid=?"
			fromid, err := strconv.Atoi(*from)
			if err != nil {
				log.Fatalf("invalid agent ID (-from=%v)", *from)
			}
			iargs = append(iargs, fromid)
		} else {
			filters[0] = "AND send.prototype=?"
			iargs = append(iargs, *from)
		}
	}
	if *to != "" {
		if *byagent {
			filters[1] = "AND t.receiverid=?"
			toid, err := strconv.Atoi(*to)
			if err != nil {
				log.Fatalf("invalid agent ID (-to=%v)", *to)
			}
			iargs = append(iargs, toid)
		} else {
			filters[1] = "AND recv.prototype=?"
			iargs = append(iargs, *to)
		}
	}
	if *commod != "" {
		filters[2] = "AND t.commodity=?"
		iargs = append(iargs, *commod)
	}
	filters[3] = nuclidefilter(*nucs)

	tmpl := template.Must(template.New("sql").Parse(s))
	var buf bytes.Buffer
	tmpl.Execute(&buf, filters)
	customSql[cmd] = buf.String()
	doCustom(os.Stdout, cmd, iargs...)
}

func doInv(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	plotit := fs.Bool("p", false, "plot the data")
	nucs := fs.String("nucs", "", "filter by comma separated `nuclide`s")
	fs.Usage = func() {
		log.Printf("Usage: %v <prototype>", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	if fs.NArg() < 1 {
		log.Fatal("must specify a prototype")
	}
	initdb()

	proto := fs.Arg(0)

	filter := nuclidefilter(*nucs)
	s := ""
	if filter != "" {
		s = `
SELECT tl.Time AS Time,IFNULL(sub.qty, 0) AS Quantity FROM timelist as tl
LEFT JOIN (
	SELECT tl.Time as time,SUM(inv.Quantity*c.MassFrac) AS qty
	FROM inventories as inv
	JOIN timelist as tl ON UNLIKELY(inv.starttime <= tl.time) AND inv.endtime > tl.time AND tl.simid=inv.simid
	JOIN agents as a on a.agentid=inv.agentid AND a.simid=inv.simid
	JOIN compositions as c on c.qualid=inv.qualid AND c.simid=inv.simid
	WHERE a.simid=? AND a.prototype=? {{.}}
	GROUP BY tl.Time
) AS sub ON sub.time=tl.time
WHERE tl.simid=?
`
	} else {
		s = `
SELECT tl.Time AS Time,IFNULL(sub.qty, 0) AS Quantity
FROM timelist as tl
LEFT JOIN (
	SELECT tl.Time as time,SUM(inv.Quantity) AS qty
	FROM inventories as inv
	JOIN timelist as tl ON UNLIKELY(inv.starttime <= tl.time) AND inv.endtime > tl.time AND tl.simid=inv.simid
	JOIN agents as a on a.agentid=inv.agentid AND a.simid=inv.simid
	WHERE a.simid=? AND a.prototype=? {{.}}
	GROUP BY tl.Time
) AS sub ON sub.time=tl.time
WHERE tl.simid=?
`
	}

	tmpl := template.Must(template.New("sql").Parse(s))
	var buf bytes.Buffer
	tmpl.Execute(&buf, filter)
	customSql[cmd] = buf.String()
	var buff bytes.Buffer
	doCustom(&buff, cmd, simid, proto, simid)
	if *plotit {
		plot(&buff, "linespoints", "Time (Months)", proto+" inventory ( kg "+*nucs+")", "Inventory")
	} else {
		fmt.Print(buff.String())
	}
}

func doFlow(cmd string, args []string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	plotit := fs.Bool("p", false, "plot the data")
	commod := fs.String("commod", "", "filter by a commodity")
	from := fs.String("from", "", "filter by supplying prototype")
	to := fs.String("to", "", "filter by receiving prototype")
	byagent := fs.Bool("byagent", false, "switch to/from filters to be agent IDs")
	nucs := fs.String("nucs", "", "filter by comma separated `nuclide`s")
	fs.Usage = func() {
		log.Printf("Usage: %v", cmd)
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	s := `
SELECT tl.Time AS Time,TOTAL(sub.qty) AS Quantity
FROM timelist as tl
LEFT JOIN (
	SELECT t.simid AS simid,t.time as time,SUM(c.massfrac*r.quantity) as qty
	FROM transactions AS t
	JOIN resources as r ON t.resourceid=r.resourceid AND r.simid=t.simid
	JOIN agents as send ON t.senderid=send.agentid AND send.simid=t.simid
	JOIN agents as recv ON t.receiverid=recv.agentid AND recv.simid=t.simid
	JOIN compositions as c ON c.qualid=r.qualid AND c.simid=r.simid
	WHERE t.simid=? {{index . 0}} {{index . 1}} {{index . 2}} {{index . 3}}
	GROUP BY t.time
) AS sub ON tl.time=sub.time AND tl.simid=sub.simid
WHERE tl.simid=?
GROUP BY tl.Time;
`

	filters := make([]string, 4)
	iargs := []interface{}{simid}
	if *from != "" {
		if *byagent {
			filters[0] = "AND t.senderid=?"
			fromid, err := strconv.Atoi(*from)
			if err != nil {
				log.Fatalf("invalid agent ID (-from=%v)", *from)
			}
			iargs = append(iargs, fromid)
		} else {
			filters[0] = "AND send.prototype=?"
			iargs = append(iargs, *from)
		}
	}
	if *to != "" {
		if *byagent {
			filters[1] = "AND t.receiverid=?"
			toid, err := strconv.Atoi(*to)
			if err != nil {
				log.Fatalf("invalid agent ID (-to=%v)", *to)
			}
			iargs = append(iargs, toid)
		} else {
			filters[1] = "AND recv.prototype=?"
			iargs = append(iargs, *to)
		}
	}
	if *commod != "" {
		filters[2] = "AND t.commodity=?"
		iargs = append(iargs, *commod)
	}
	filters[3] = " " + nuclidefilter(*nucs)
	iargs = append(iargs, simid)

	tmpl := template.Must(template.New("sql").Parse(s))
	var buf bytes.Buffer
	tmpl.Execute(&buf, filters)
	customSql[cmd] = buf.String()
	var buff bytes.Buffer
	doCustom(&buff, cmd, iargs...)
	if *plotit {
		plot(&buff, "impulses", "Time (Months)", "Quantity Transacted ( kg "+*nucs+")", "Flow")
	} else {
		fmt.Print(buff.String())
	}
}

func doFlowGraph(cmd string, args []string) {
	fs := flag.NewFlagSet("flowgraph", flag.ExitOnError)
	fs.Usage = func() {
		log.Print("Usage: flowgraph")
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	proto := fs.Bool("proto", false, "aggregate nodes by prototype")
	t0 := fs.Int("t1", 0, "beginning of time interval (default is beginning of simulation)")
	t1 := fs.Int("t2", -1, "end of time interval (default if end of simulation)")
	fs.Parse(args)
	initdb()

	arcs, err := query.FlowGraph(db, simid, *t0, *t1, *proto)
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

func doCreated(cmd string, args []string) {
	fs := flag.NewFlagSet("created", flag.ExitOnError)
	fs.Usage = func() {
		log.Print("Usage: created [agent-id...]\nZero agents uses all agents")
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	t0 := fs.Int("t1", 0, "beginning of time interval (default is beginning of simulation)")
	t1 := fs.Int("t2", -1, "end of time interval (default if end of simulation)")
	fs.Parse(args)
	initdb()

	var agents []int

	for _, arg := range fs.Args() {
		id, err := strconv.Atoi(arg)
		fatalif(err)
		agents = append(agents, id)
	}

	m, err := query.MatCreated(db, simid, *t0, *t1, agents...)
	fatalif(err)
	fmt.Printf("%+v\n", m)
}

func doEnergy(cmd string, args []string) {
	fs := flag.NewFlagSet("energy", flag.ExitOnError)
	t0 := fs.Int("t1", 0, "beginning of time interval (default is beginning of simulation)")
	t1 := fs.Int("t2", -1, "end of time interval (default if end of simulation)")
	fs.Usage = func() {
		log.Print("Usage: energy")
		log.Printf("%v\n", cmds.Help(cmd))
		fs.PrintDefaults()
	}
	fs.Parse(args)
	initdb()

	e, err := query.EnergyProduced(db, simid, *t0, *t1)
	fatalif(err)
	fmt.Println(e)
}

func fatalif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type CmdSet struct {
	funcs map[string]func(string, []string) // map[cmdname]func(cmdname, args)
	Names []string
	Helps []string
}

func NewCmdSet() *CmdSet {
	return &CmdSet{funcs: map[string]func(string, []string){}}
}

func (cs *CmdSet) IsDiv(i int) bool {
	_, ok := cs.funcs[cs.Names[i]]
	return !ok
}

func (cs *CmdSet) Help(cmd string) string {
	for i := range cs.Names {
		if cs.Names[i] == cmd {
			return cs.Helps[i]
		}
	}
	return ""
}

func (cs *CmdSet) RegisterDiv(name string) {
	cs.Names = append(cs.Names, name)
	cs.Helps = append(cs.Helps, "")
}

func (cs *CmdSet) Register(name, brief string, f func(string, []string)) {
	cs.Names = append(cs.Names, name)
	cs.Helps = append(cs.Helps, brief)
	cs.funcs[name] = f
}

func (cs *CmdSet) Execute(args []string) {
	cmd := args[0]
	f, ok := cs.funcs[cmd]
	if !ok {
		blankargs := make([]interface{}, len(args)-1)
		for i, arg := range args[1:] {
			blankargs[i] = arg
		}
		initdb()
		doCustom(os.Stdout, cmd, blankargs...)
		return
	}
	f(cmd, args[1:])
}

func nuclidefilter(nucs string) string {
	if len(nucs) == 0 {
		return ""
	}

	nnucs := []nuc.Nuc{}
	for _, n := range strings.Split(nucs, ",") {
		nuc, err := nuc.Id(strings.TrimSpace(n))
		fatalif(err)
		nnucs = append(nnucs, nuc)
	}

	if len(nnucs) == 1 {
		return fmt.Sprintf(" AND c.nucid = %v", int(nnucs[0]))
	}

	filter := fmt.Sprintf(" AND c.nucid IN (%v", int(nnucs[0]))
	for _, nuc := range nnucs[1:] {
		filter += fmt.Sprintf(",%v", int(nuc))
	}
	return filter + ") "
}
