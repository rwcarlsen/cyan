package taint

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type Visited map[*Node]struct{}

type Node struct {
	Id        int
	ResId     int
	AgentId   int
	Time      int
	QualId    int
	Quantity  float64
	Parent1   *Node
	Parent2   *Node
	Child1    *Node
	Child2    *Node
	taintfrac float64
	par1mark  bool
	par2mark  bool
}

type bytime []*NodeData

func (ns bytime) Len() int      { return len(ns) }
func (ns bytime) Swap(i, j int) { ns[i], ns[j] = ns[j], ns[i] }
func (ns bytime) Less(i, j int) bool {
	if ns[i].Time != ns[j].Time {
		return ns[i].Time < ns[j].Time
	} else if ns[i].ResId != ns[j].ResId {
		return ns[i].ResId < ns[j].ResId
	} else if ns[i].AgentId != -1 && ns[j].AgentId != -1 {
		return ns[i].AgentId < ns[j].AgentId
	}
	return false
}

func (n *Node) String() string {
	var buf bytes.Buffer
	n.str(&buf, 0)
	return buf.String()
}

func (n *Node) str(buf *bytes.Buffer, indent int) {
	if n == nil {
		fmt.Fprintf(buf, "<nil>")
	} else {
		ind := strings.Repeat(" ", indent)
		fmt.Fprintf(buf, "{\n")
		fmt.Fprintf(buf, ind+"    ResId:   %v,\n", n.ResId)
		fmt.Fprintf(buf, ind+"    AgentId: %v,\n", n.AgentId)
		fmt.Fprintf(buf, ind+"    Time:    %v,\n", n.Time)
		fmt.Fprintf(buf, ind+"    Child1:  ")
		n.Child1.str(buf, indent+4)
		fmt.Fprintf(buf, ind+"    Child2:  ")
		n.Child2.str(buf, indent+4)
		fmt.Fprintf(buf, strings.Repeat(" ", indent)+"}")
	}

	if indent > 0 {
		fmt.Fprintf(buf, ",")
	}
	fmt.Fprintf(buf, "\n")
}

type NodeData struct {
	ResId    int
	AgentId  int
	QualId   int
	Time     int
	Quantity float64
	Parent1  int
	Parent2  int
}

// nodemap is map[resid][]Node* (by location/agent)
type nodeMap map[int][]*Node

func Tree(nodes []*NodeData) (roots []*Node) {
	sort.Sort(bytime(nodes))
	nodemap := nodeMap{}
	nextid := 0

	for _, row := range nodes {
		node := &Node{
			Id:       nextid,
			AgentId:  row.AgentId,
			ResId:    row.ResId,
			Time:     row.Time,
			Quantity: row.Quantity,
			QualId:   row.QualId,
		}
		nextid++

		parent1, parent2 := row.Parent1, row.Parent2
		if len(nodemap[node.ResId]) == 0 {

			// this node's parent(s) have a different resource id than this node
			if parent1 != 0 {
				p1 := nodemap[parent1][len(nodemap[parent1])-1]
				node.Parent1 = p1
				if p1.Child1 == nil {
					p1.Child1 = node
				} else {
					p1.Child2 = node
				}
			}
			if parent2 != 0 {
				p2 := nodemap[parent2][len(nodemap[parent2])-1]
				node.Parent2 = p2
				if p2.Child1 == nil {
					p2.Child1 = node
				} else {
					p2.Child2 = node
				}
			}
		} else { // there is already a node with this resource id
			parent := nodemap[node.ResId][len(nodemap[node.ResId])-1]
			node.Parent1 = parent
			parent.Child1 = node
		}

		nodemap[node.ResId] = append(nodemap[node.ResId], node)
		if parent1 == 0 && parent2 == 0 {
			roots = append(roots, node)
		}
	}

	v := Visited{}
	for _, root := range roots {
		root.fixagentid(v)
	}
	return roots
}

// fixagentid assigns missing AgentId's to nodes. some nodes if generated from a cyclus database query don't have AgentId's
// associated with them and had them marked as -1.  These are ommitted from
// the db because they don't affect inventories (i.e. intra-time-step,
// intra-agent modifications).  So we can fill these in by assigning the same
// agentid as the parent node(s).
func (n *Node) fixagentid(visited Visited) {
	if n == nil {
		return
	} else if _, ok := visited[n]; ok {
		return
	}
	visited[n] = struct{}{}

	if n.AgentId == -1 {
		if n.Parent1 != nil {
			n.AgentId = n.Parent1.AgentId
		} else {
			// give up?
		}
	}

	n.Child1.fixagentid(visited)
	n.Child2.fixagentid(visited)
}

type TaintVal struct {
	Taint    float64
	Quantity float64
}

// Locate searches for and returns the neares (shallowest) node with the given
// Resource ID (resid).  It returns nil if not found.
func (n *Node) Locate(v Visited, resid int) *Node {
	if n == nil {
		return nil
	} else if _, ok := v[n]; ok {
		return nil
	}
	v[n] = struct{}{}

	if n.ResId == resid {
		return n
	} else if got := n.Child1.Locate(v, resid); got != nil {
		return got
	} else if got := n.Child2.Locate(v, resid); got != nil {
		return got
	}
	return nil
}

func (n *Node) ResetTaint(v Visited) {
	if n == nil {
		return
	} else if _, ok := v[n]; ok {
		return
	}
	v[n] = struct{}{}

	n.taintfrac = -1
	n.par1mark = false
	n.par2mark = false
	n.Child1.ResetTaint(v)
	n.Child2.ResetTaint(v)
}

// Taint returns a map of agent ID to a slice/time-series of taint values of
// aggregate resource in that agent originating from the node's resource
// object going forward down the graph through all time.
func (n *Node) Taint(tmax int) map[int][]TaintVal {
	all := map[int][]TaintVal{}
	n.ResetTaint(Visited{})

	n.taintfrac = 1.0

	// mark dirty edges
	v := Visited{}
	n.Child1.mark(v)
	n.Child2.mark(v)

	// calculate taintfracs
	v = Visited{}
	n.Child1.taint(v)
	n.Child2.taint(v)

	// aggregate by agent id and time
	n.taintnodes(all, tmax)

	return all
}

func (n *Node) marked() bool {
	return (n != nil) && (n.par1mark || n.par2mark)
}

func (n *Node) mark(v Visited) {
	if n == nil {
		return
	} else if _, ok := v[n]; ok {
		return
	}
	v[n] = struct{}{}

	if n.Parent1.marked() {
		n.par1mark = true
	}
	if n.Parent2.marked() {
		n.par2mark = true
	}

	n.Child1.mark(v)
	n.Child2.mark(v)
}

// taint calculates the taint on each node using a recursive depth-first walk.
func (n *Node) taint(v Visited) {
	if n == nil {
		return
	} else if _, ok := v[n]; ok {
		return
	}

	// make sure potential tainted parent nodes have been calculated before
	// calculating taint for this node.
	if n.par1mark && n.Parent1.taintfrac < 0 {
		return
	} else if n.par2mark && n.Parent2.taintfrac < 0 {
		return
	}

	v[n] = struct{}{}

	if n.Parent2 == nil { // from a transmute, move, split
		n.taintfrac = n.Parent1.taintfrac
	} else { // from a combine/absorb
		n.taintfrac = (n.Parent1.taintfrac*n.Parent1.Quantity +
			n.Parent2.taintfrac*n.Parent2.Quantity) / n.Quantity
	}

	n.Child1.taint(v)
	n.Child2.taint(v)
}

// taintnodes walks the tree building a time-series of taint values for each
// agent id.
func (n *Node) taintnodes(all map[int][]TaintVal, tmax int) {
	for len(all[n.AgentId]) < tmax {
		all[n.AgentId] = append(all[n.AgentId], TaintVal{})
	}

	torec := false
	torec = torec || (n.Child1 == nil && n.Child2 == nil)
	torec = torec || (n.Child1 != nil && n.Time != n.Child1.Time)

	if torec {
		prev := all[n.AgentId][n.Time]
		qty := prev.Quantity + n.Quantity
		taintqty := prev.Taint*prev.Quantity + n.taintfrac*n.Quantity
		all[n.AgentId][n.Time] = TaintVal{
			Taint:    taintqty / qty,
			Quantity: qty,
		}

		// fill in blank times between this node and its next child
		if n.Child1 != nil {
			for t := n.Time + 1; t < n.Child1.Time; t++ {
				prev := all[n.AgentId][t]
				qty := prev.Quantity + n.Quantity
				taintqty := prev.Taint*prev.Quantity + n.taintfrac*n.Quantity
				all[n.AgentId][t] = TaintVal{
					Taint:    taintqty / qty,
					Quantity: qty,
				}
			}
		} else if n.Child1 == nil {
			// leaf node taint needs to be forward propogated through all blank times
			for i, prev := range all[n.AgentId][n.Time+1:] {
				t := i + n.Time + 1
				qty := prev.Quantity + n.Quantity
				taintqty := prev.Taint*prev.Quantity + n.taintfrac*n.Quantity
				all[n.AgentId][t] = TaintVal{
					Taint:    taintqty / qty,
					Quantity: qty,
				}
			}
		}
	}

	if n.Child1 != nil {
		n.Child1.taintnodes(all, tmax)
	}
	if n.Child2 != nil {
		n.Child2.taintnodes(all, tmax)
	}
}

func TreeFromDb(db *sql.DB, simid []byte) (roots []*Node) {
	s := `SELECT r.ResourceId,r.TimeCreated,inv.StartTime,r.Quantity,r.QualId,r.Parent1,r.Parent2,inv.AgentId
	      FROM resources AS r
		  LEFT JOIN Inventories AS inv ON inv.SimId = r.SimId AND inv.ResourceId = r.ResourceId
		  WHERE r.SimId = ?
		  ORDER BY r.ResourceId,r.TimeCreated,inv.StartTime`
	rows, err := db.Query(s, simid)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	nodes := []*NodeData{}
	var time2 sql.NullInt64
	var agentid sql.NullInt64

	for rows.Next() {
		n := &NodeData{}

		err := rows.Scan(&n.ResId, &n.Time, &time2, &n.Quantity, &n.QualId, &n.Parent1, &n.Parent2, &agentid)
		if err != nil {
			panic(err.Error())
		}

		// if the resource object moved to a new agent after being created in
		// its current state, then the node must be associated with the time
		// when the resource moved into that agent rather than when it was
		// created.
		if time2.Valid && int(time2.Int64) > n.Time {
			n.Time = int(time2.Int64)
		}

		n.AgentId = -1
		if agentid.Valid {
			n.AgentId = int(agentid.Int64)
		}

		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		panic(err.Error())
	}

	return Tree(nodes)
}
