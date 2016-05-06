package taint

import "testing"

var taintcases = []struct {
	Descrip   string
	TreeIndex int
	Res       int
	Want      map[int][]TaintVal
}{
	{
		Descrip:   "split spans the time step so splits' parent should not be double counted in taint analysis",
		TreeIndex: 1,
		Res:       1,
		Want: map[int][]TaintVal{
			1: []TaintVal{
				{Taint: 0.0, Quantity: 0},
				{Taint: 1.0, Quantity: 3},
				{Taint: 1.0, Quantity: 2},
			},
			2: []TaintVal{
				{Taint: 0.0, Quantity: 0},
				{Taint: 0.0, Quantity: 0},
				{Taint: 1.0, Quantity: 1},
			},
		},
	},
}

var treecases = []struct {
	Descrip string
	Note    string
	Raw     []*NodeData
	Want    []*Node
}{
	{
		Descrip: "agent-internal single split",
		Raw: []*NodeData{
			{ResId: 1, AgentId: 1, Parent1: 0, Parent2: 0},
			{ResId: 2, AgentId: 1, Parent1: 1, Parent2: 0},
			{ResId: 3, AgentId: 1, Parent1: 1, Parent2: 0},
		},
		Want: []*Node{
			{
				ResId:   1,
				AgentId: 1,
				Child1: &Node{
					ResId:   2,
					AgentId: 1,
				},
				Child2: &Node{
					ResId:   3,
					AgentId: 1,
				},
			},
		},
	}, {
		Descrip: "agent-internal split then move",
		Raw: []*NodeData{
			{ResId: 1, AgentId: 1, Parent1: 0, Parent2: 0, Time: 1, Quantity: 3},
			{ResId: 2, AgentId: 1, Parent1: 1, Parent2: 0, Time: 2, Quantity: 1},
			{ResId: 3, AgentId: 1, Parent1: 1, Parent2: 0, Time: 2, Quantity: 2},
			{ResId: 2, AgentId: 2, Parent1: 1, Parent2: 0, Time: 2, Quantity: 1},
		},
		Want: []*Node{
			{
				ResId:   1,
				AgentId: 1,
				Child1: &Node{
					ResId:   2,
					AgentId: 1,
					Child1: &Node{
						ResId:   2,
						AgentId: 2,
					},
				},
				Child2: &Node{
					ResId:   3,
					AgentId: 1,
				},
			},
		},
	}, {
		Descrip: "multiple roots, no mods",
		Raw: []*NodeData{
			{ResId: 1, AgentId: 1, Parent1: 0, Parent2: 0},
			{ResId: 2, AgentId: 1, Parent1: 0, Parent2: 0},
		},
		Want: []*Node{
			{
				ResId:   1,
				AgentId: 1,
			}, {
				ResId:   2,
				AgentId: 1,
			},
		},
	},
}

func TestTree(t *testing.T) {
	for i, test := range treecases {
		tree := Tree(test.Raw)
		for j, wantroot := range test.Want {
			if !deepequal(tree[j], wantroot) || !validconn(tree[j]) {
				t.Errorf("test %v (%v), root %v failed - want:\n %v", i+1, test.Descrip, j+1, wantroot)
			}
			t.Logf("test %v (%v), root %v - got:\n%v", i+1, test.Descrip, j+1, tree[j])
		}
	}
}

func TestNode_Taint(t *testing.T) {
	for i, test := range taintcases {
		t.Logf("Case %v (tree %v, res %v)", i, test.TreeIndex+1, test.Res)
		roots := Tree(treecases[test.TreeIndex].Raw)

		var tree *Node
		for _, root := range roots {
			tree = root.Locate(test.Res)
			if tree != nil {
				break
			}
		}
		if tree == nil {
			t.Errorf("  FAIL: could not locate resource")
			continue
		}

		taints := tree.Taint()
		for agentid, got := range taints {
			t.Logf("  Agent %v:", agentid)
			want := test.Want[agentid]
			if len(want) != len(got) {
				t.Errorf("    FAIL: got %v, want %v", got, want)
				continue
			}
			for j := range want {
				if got[j] != want[j] {
					t.Errorf("    FAIL t=%v: got %+v, want %+v", j, got[j], want[j])
					break
				} else {
					t.Logf("         t=%v: got %+v", j, got[j])
				}
			}
		}
	}
}

// TestNode_String just checks that the String function doesn't panic
func TestNode_String(t *testing.T) {
	for _, test := range treecases {
		tree := Tree(test.Raw)
		tree[0].String()
	}
}

func validconn(node *Node) bool {
	if node == nil {
		return true
	}

	if node.Parent1 != nil {
		if node.Parent1.Child1 != node && node.Parent1.Child2 != node {
			return false
		}

	}
	if node.Parent2 != nil {
		if node.Parent2.Child1 != node && node.Parent2.Child2 != node {
			return false
		}

	}

	if !validconn(node.Child1) {
		return false
	}
	if !validconn(node.Child2) {
		return false
	}
	return true
}

func deepequal(tree1, tree2 *Node) bool {
	if tree1 == nil || tree2 == nil {
		return tree1 == tree2
	}

	if tree1.ResId != tree2.ResId {
		return false
	} else if tree1.AgentId != tree2.AgentId {
		return false
	}

	if !deepequal(tree1.Child1, tree2.Child1) {
		return false
	} else if !deepequal(tree1.Child2, tree2.Child2) {
		return false
	}
	return true
}
