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

	"code.google.com/p/go-uuid/uuid"
	"github.com/mxk/go-sqlite/sqlite3"
	"github.com/rwcarlsen/cyan/post"
	"github.com/rwcarlsen/cyan/query"
)

const MAX_MEMORY = 50 * 1024 * 1024

var resultTmpl = template.Must(template.New("results").Parse(results))

var addr = flag.String("addr", "127.0.0.1:4141", "network address of dispatch server")

func main() {
	flag.Parse()
	s := NewServer()
	s.ListenAndServe(*addr)
	if err := s.ListenAndServe(*addr); err != nil {
		log.Fatal(err)
	}
}

type Server struct {
	Dbs map[string]string
}

func NewServer() *Server {
	return &Server{
		Dbs: map[string]string{},
	}
}

func (s *Server) ListenAndServe(addr string) error {
	http.HandleFunc("/", s.main)
	http.HandleFunc("/upload", s.upload)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) main(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(home))
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	// parse database from multi part form data
	if err := r.ParseMultipartForm(MAX_MEMORY); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	gotDb := false
	fname := uuid.NewRandom().String() + ".sqlite"
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

	// post process the database
	conn, err := sqlite3.Open(fname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer conn.Close()

	err = post.Prepare(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer post.Finish(conn)

	simids, err := post.GetSimIds(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	ctx := post.NewContext(conn, simids[0], nil)
	if err := ctx.WalkAll(); err != nil {
		fmt.Println(err)
	}

	// get simid
	db, err := sql.Open("sqlite3", fname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer db.Close()

	ids, err := query.SimIds(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	simid := ids[0]
	rs := &Results{}

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

	// create agents table
	rs.Agents, err = query.AllAgents(db, simid, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	// create transactions table
	sql := `
		SELECT TransactionId,Time,SenderId,ReceiverId,tr.ResourceId,Commodity,NucId,MassFrac*Quantity
		FROM Transactions AS tr
		INNER JOIN Resources AS res ON tr.ResourceId = res.ResourceId
		INNER JOIN Compositions AS cmp ON res.QualId = cmp.QualId
		WHERE tr.SimId = ? AND cmp.SimId = res.SimId AND res.SimId = tr.SimId;
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
		rs.Trans = append(rs.Trans, t)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}

	// render all rsults
	resultTmpl.Execute(w, rs)
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
	Flowgraph string
	Agents    []query.AgentInfo
	Trans     []*Trans
}

//<input type="hidden" name="redirect" value="/analysis" />

const home = `
<html>
<head>
	<meta charset="UTF-8"/>
</head>
<body>
	<h1>Cyclus Data Viewer</h1>

	<form action="/upload" method="POST" enctype="multipart/form-data">

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
	<meta charset="UTF-8"/>
</head>
<body>

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
	
	<h2>Transactions Table</h2>
	<table>
		<tr><th>ID</th><th>Time</th><th>Sender</th><th>Receiver</th><th>ResourceId</th><th>Commod</th><th>Nuclide</th><th>Quantity (kg)</th></tr>

		{{ range .Trans}}
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

</body>
</html>
`
