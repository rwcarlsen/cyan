package taint

import "testing"

var cases = []struct {
	Raw  []*NodeData
	Want []*Node
}{
	{
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
		Raw: []*NodeData{
			{ResId: 1, AgentId: 1, Parent1: 0, Parent2: 0},
			{ResId: 2, AgentId: 1, Parent1: 1, Parent2: 0},
			{ResId: 2, AgentId: 2, Parent1: 1, Parent2: 0},
			{ResId: 3, AgentId: 1, Parent1: 1, Parent2: 0},
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
	for i, test := range cases {
		tree := Tree(test.Raw)
		for j, wantroot := range test.Want {
			if !deepequal(tree[j], wantroot) || !validconn(tree[j]) {
				t.Errorf("test %v, root %v failed - want:\n %v", i+1, j+1, wantroot)
			}
			t.Logf("test %v, root %v - got:\n%v", i+1, j+1, tree[j])
		}
	}
}

// TestNode_String just checks that the String function doesn't panic
func TestNode_String(t *testing.T) {
	for _, test := range cases {
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
