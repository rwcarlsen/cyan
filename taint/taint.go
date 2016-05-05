package taint

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
)

type Node struct {
	Id       int
	ResId    int
	AgentId  int
	Time     int
	QualId   int
	Quantity float64
	Parent1  *Node
	Parent2  *Node
	Child1   *Node
	Child2   *Node
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
		if node.Parent1 == nil && node.Parent2 == nil {
			roots = append(roots, node)
		}
	}
	return roots
}

func TreeFromDb(db *sql.DB, simid []byte) (roots []*Node) {
	s := `SELECT r.ResourceId,r.TimeCreated,inv.StartTime,r.Quantity,r.QualId,r.Parent1,r.Parent2,inv.AgentId
	      FROM resources AS r
		  INNER JOIN Inventories AS inv ON inv.SimId = r.SimId AND inv.ResourceId = r.ResourceId
		  WHERE r.SimId = ?
		  ORDER BY r.TimeCreated,r.ResourceId,inv.StartTime`
	rows, err := db.Query(s, simid)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	nodes := []*NodeData{}
	var time2 int

	for rows.Next() {
		n := &NodeData{}

		err := rows.Scan(&n.ResId, &n.Time, &time2, &n.Quantity, &n.QualId, &n.Parent1, &n.Parent2, &n.AgentId)
		if err != nil {
			panic(err.Error())
		}

		// if the resource object moved to a new agent after being created in
		// its current state, then the node must be associated with the time
		// when the resource moved into that agent rather than when it was
		// created.
		if time2 > n.Time {
			n.Time = time2
		}

		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		panic(err.Error())
	}

	return Tree(nodes)
}
