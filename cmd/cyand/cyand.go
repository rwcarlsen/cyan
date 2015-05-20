package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/rwcarlsen/cyan/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	_ "github.com/rwcarlsen/cyan/Godeps/_workspace/src/github.com/mxk/go-sqlite/sqlite3"
	"github.com/rwcarlsen/cyan/post"
	"github.com/rwcarlsen/cyan/query"
)

const MAX_MEMORY = 50 * 1024 * 1024
const timeout = 30 * time.Second

var resultTmpl = template.Must(template.New("results").Parse(results))
var homeTmpl = template.Must(template.New("home").Parse(home))

var addr = flag.String("addr", "127.0.0.1:4141", "network address of dispatch server")

func main() {
	flag.Parse()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/upload/", upload)
	http.HandleFunc("/share/", share)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	uid := uuid.NewRandom().String()
	homeTmpl.Execute(w, uid)
}

func share(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Path[len("/share/"):]
	fname := uid + ".html"
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	w.Write(data)
}

func uploadInner(w http.ResponseWriter, r *http.Request, kill chan bool) {
	// parse database from multi part form data
	if err := r.ParseMultipartForm(MAX_MEMORY); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	gotDb := false
	uid := r.URL.Path[len("/upload/"):]
	fname := uid + ".sqlite"
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, _ := fileHeader.Open()
			data, _ := ioutil.ReadAll(file)
			err := ioutil.WriteFile(fname, data, 0644)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Print(err)
				return
			}
			defer func() { os.Remove(fname) }()
			gotDb = true
			break
		}
		break
	}
	if !gotDb {
		http.Error(w, "No file provided", http.StatusBadRequest)
		log.Print("received request with no file")
		return
	}

	select {
	case <-kill:
		return
	default:
	}

	// post process the database
	db, err := sql.Open("sqlite3", fname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer db.Close()

	err = post.Prepare(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer post.Finish(db)

	simids, err := post.GetSimIds(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	ctx := post.NewContext(db, simids[0])
	if err := ctx.WalkAll(); err != nil {
		log.Println(err)
	}

	// get simid
	ids, err := query.SimIds(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	simid := ids[0]
	rs := &Results{}

	select {
	case <-kill:
		return
	default:
	}

	// create flow graph
	combineProto := false
	arcs, err := query.FlowGraph(db, simid, 0, -1, combineProto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	dotname := fname + ".dot"
	dotf, err := os.Create(dotname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer func() { os.Remove(dotname) }()

	fmt.Fprintln(dotf, "digraph ResourceFlows {")
	fmt.Fprintln(dotf, "    overlap = false;")
	fmt.Fprintln(dotf, "    nodesep=1.0;")
	fmt.Fprintln(dotf, "    edge [fontsize=9];")
	for _, arc := range arcs {
		fmt.Fprintf(dotf, "    \"%v\" -> \"%v\" [label=\"%v\\n(%.3g kg)\"];\n", arc.Src, arc.Dst, arc.Commod, arc.Quantity)
	}
	fmt.Fprintln(dotf, "}")
	dotf.Close()

	var buf bytes.Buffer
	cmd := exec.Command("dot", "-Tsvg", dotname)
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	rs.Flowgraph = buf.String()

	select {
	case <-kill:
		return
	default:
	}

	// create agents table
	rs.Agents, err = query.AllAgents(db, simid, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	// create Material transactions table
	sql := `
		SELECT TransactionId,Time,SenderId,ReceiverId,tr.ResourceId,Commodity,NucId,MassFrac*Quantity
		FROM Transactions AS tr
		INNER JOIN Resources AS res ON tr.ResourceId = res.ResourceId
		INNER JOIN Compositions AS cmp ON res.QualId = cmp.QualId
		WHERE tr.SimId = ? AND cmp.SimId = res.SimId AND res.SimId = tr.SimId
		AND res.Type = 'Material';
		`
	rows, err := db.Query(sql, simid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	for rows.Next() {
		t := &Trans{}
		if err := rows.Scan(&t.Id, &t.Time, &t.Sender, &t.Receiver, &t.ResourceId, &t.Commod, &t.Nuc, &t.Qty); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Print(err)
			return
		}
		rs.TransMats = append(rs.TransMats, t)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	// create Material transactions table
	sql = `
		SELECT TransactionId,Time,SenderId,ReceiverId,tr.ResourceId,Commodity,Quality,Quantity
		FROM Transactions AS tr
		INNER JOIN Resources AS res ON tr.ResourceId = res.ResourceId
		INNER JOIN Products AS pd ON res.QualId = pd.QualId
		WHERE tr.SimId = ? AND pd.SimId = res.SimId AND res.SimId = tr.SimId
		AND res.Type = 'Product';
		`
	rows, err = db.Query(sql, simid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	for rows.Next() {
		t := &ProdTrans{}
		if err := rows.Scan(&t.Id, &t.Time, &t.Sender, &t.Receiver, &t.ResourceId, &t.Commod, &t.Quality, &t.Qty); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Print(err)
			return
		}
		rs.TransProds = append(rs.TransProds, t)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	select {
	case <-kill:
		return
	default:
	}

	// render all results and save page
	rs.Uid = uid
	resultTmpl.Execute(w, rs)

	f, err := os.Create(uid + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer f.Close()
	resultTmpl.Execute(f, rs)
	close(kill)
}

func upload(w http.ResponseWriter, r *http.Request) {
	kill := make(chan bool)
	go uploadInner(w, r, kill)

	select {
	case <-time.After(timeout):
		close(kill)
		http.Error(w, "operation timed out", http.StatusInternalServerError)
		log.Print("db processing operation timed out")
	case <-kill:
	}
}

type ProdTrans struct {
	Id         int
	Time       int
	Sender     int
	Receiver   int
	ResourceId int
	Commod     string
	Quality    string
	Qty        float64
}

type Trans struct {
	Id         int
	Time       int
	Sender     int
	Receiver   int
	ResourceId int
	Commod     string
	Nuc        int
	Qty        float64
}

type Results struct {
	Uid        string
	Flowgraph  string
	Agents     []query.AgentInfo
	TransMats  []*Trans
	TransProds []*ProdTrans
}

//<input type="hidden" name="redirect" value="/analysis" />

const home = `
<html>
<head>
	<meta charset="UTF-8"/>
</head>
<body>

	<h1>Cyclus Data Viewer</h1>

	<form action="/upload/{{.}}" method="POST" enctype="multipart/form-data">
		<label for="file">Cyclus Sqlite Database:</label>
		<input name="file" type="file"></input>
		<input type="submit"></input>
	</form>

</body>
</html>
`

const results = `
<html>
<head>
	<title>Cyclus Database Viewer</title>
	<meta charset="UTF-8"/>
	<style>
		h2 {
			background-color:#8AC1EC;
			text-align:center;
		}
		table {
			width:80%;
			border-color:#a9a9a9;
			color:#333333;
			border-collapse:collapse;
			margin:auto;
			border-width:1px;
			text-align:center;
		}
		th {
			padding:4px;
			border-style:solid;
			border-color:#a9a9a9;
			border-width:1px;
			background-color:#b8b8b8;
			text-align:left;
		}
		tr {
			background-color:#ffffff;
			text-align:center;
		}
		td {
			padding:4px;
			border-color:#a9a9a9;
			border-style:solid;
			border-width:1px;
			text-align:center;
		}
	</style>
</head>
<body>

    <h3>Share link: <a href="/share/{{.Uid}}">http://cyc-submit.rwcr.net/share/{{.Uid}}</a></h3>

	<h2>Resource Flow</h2>
	{{.Flowgraph}}

	<br><br>

	<h2>Agents Table</h2>
	<table>
		<tr><th>ID</th><th>Kind</th><th>Spec</th><th>Prototype</th><th>ParentId</th><th>Lifetime</th><th>EnterTime</th><th>ExitTime</th></tr>

		{{ range .Agents}}
		<tr>
			<td>{{ .Id }}</td>
			<td>{{ .Kind }}</td>
			<td>{{ .Impl }}</td>
			<td>{{ .Proto }}</td>
			<td>{{ .Parent }}</td>
			<td>{{ .Lifetime }}</td>
			<td>{{ .Enter }}</td>
			<td>{{ .Exit }}</td>
		</tr>
		{{ end }}
	</table>
	
	<h2>Material Transactions Table</h2>
	<table>
		<tr><th>ID</th><th>Time</th><th>Sender</th><th>Receiver</th><th>ResourceId</th><th>Commod</th><th>Nuclide</th><th>Quantity (kg)</th></tr>

		{{ range .TransMats}}
		<tr>
			<td>{{ .Id }}</td>
			<td>{{ .Time }}</td>
			<td>{{ .Sender }}</td>
			<td>{{ .Receiver }}</td>
			<td>{{ .ResourceId }}</td>
			<td>{{ .Commod }}</td>
			<td>{{ .Nuc }}</td>
			<td>{{ .Qty }}</td>
		</tr>
		{{ end }}
	</table>

	<h2>Product Transactions Table</h2>
	<table>
		<tr><th>ID</th><th>Time</th><th>Sender</th><th>Receiver</th><th>ResourceId</th><th>Commod</th><th>Quality</th><th>Quantity</th></tr>

		{{ range .TransProds}}
		<tr>
			<td>{{ .Id }}</td>
			<td>{{ .Time }}</td>
			<td>{{ .Sender }}</td>
			<td>{{ .Receiver }}</td>
			<td>{{ .ResourceId }}</td>
			<td>{{ .Commod }}</td>
			<td>{{ .Quality }}</td>
			<td>{{ .Qty }}</td>
		</tr>
		{{ end }}
	</table>

</body>
</html>
`
