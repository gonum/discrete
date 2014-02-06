package concrete_test

import (
	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"math"
	"sort"
	"testing"
)

var _ graph.DirectedGraph = &concrete.DenseGraph{}
var _ graph.CrunchGraph = &concrete.DenseGraph{}

func TestBasicDenseImpassable(t *testing.T) {
	dg := concrete.NewDenseGraph(5, false)
	if dg == nil {
		t.Fatal("Directed graph could not be made")
	}

	for i := 0; i < 5; i++ {
		if !dg.NodeExists(concrete.GonumNode(i)) {
			t.Errorf("Node that should exist doesn't: %d", i)
		}

		if degree := dg.Degree(concrete.GonumNode(i)); degree != 0 {
			t.Errorf("Node in impassable graph has a neighbor. Node: %d Degree: %d", i, degree)
		}
	}

	for i := 5; i < 10; i++ {
		if dg.NodeExists(concrete.GonumNode(i)) {
			t.Errorf("Node exists that shouldn't: %d")
		}
	}
}

func TestBasicDensePassable(t *testing.T) {
	dg := concrete.NewDenseGraph(5, true)
	if dg == nil {
		t.Fatal("Directed graph could not be made")
	}

	for i := 0; i < 5; i++ {
		if !dg.NodeExists(concrete.GonumNode(i)) {
			t.Errorf("Node that should exist doesn't: %d", i)
		}

		if degree := dg.Degree(concrete.GonumNode(i)); degree != 10 {
			t.Errorf("Node in impassable graph has a neighbor. Node: %d Degree: %d", i, degree)
		}
	}

	for i := 5; i < 10; i++ {
		if dg.NodeExists(concrete.GonumNode(i)) {
			t.Errorf("Node exists that shouldn't: %d")
		}
	}
}

func TestDenseAddRemove(t *testing.T) {
	dg := concrete.NewDenseGraph(10, false)
	dg.SetEdgeCost(concrete.GonumNode(0), concrete.GonumNode(2), 1.0, false)

	if neighbors := dg.Neighbors(concrete.GonumNode(0)); len(neighbors) != 1 || neighbors[0].ID() != 2 ||
		!dg.IsNeighbor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Errorf("Couldn't add neighbor")
	}

	if neighbors := dg.Successors(concrete.GonumNode(0)); len(neighbors) != 1 || neighbors[0].ID() != 2 ||
		!dg.IsSuccessor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Errorf("Adding edge didn't create successor")
	}

	if neighbors := dg.Predecessors(concrete.GonumNode(0)); len(neighbors) != 1 || neighbors[0].ID() != 2 ||
		!dg.IsPredecessor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Errorf("Adding undirected edge didn't create predecessor")
	}

	if neighbors := dg.Neighbors(concrete.GonumNode(2)); len(neighbors) != 1 || neighbors[0].ID() != 0 ||
		!dg.IsNeighbor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Errorf("Adding an undirected neighbor didn't add it reciprocally")
	}

	if neighbors := dg.Successors(concrete.GonumNode(2)); len(neighbors) != 1 || neighbors[0].ID() != 0 ||
		!dg.IsSuccessor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Errorf("Adding undirected edge didn't create proper successor")
	}

	if neighbors := dg.Predecessors(concrete.GonumNode(2)); len(neighbors) != 1 || neighbors[0].ID() != 0 ||
		!dg.IsPredecessor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Errorf("Adding edge didn't create proper predecessor")
	}

	dg.RemoveEdge(concrete.GonumNode(0), concrete.GonumNode(2), true)

	if neighbors := dg.Neighbors(concrete.GonumNode(0)); len(neighbors) != 1 || neighbors[0].ID() != 2 ||
		!dg.IsNeighbor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Errorf("Removing a directed edge changed result of neighbors when neighbors is undirected; neighbors: %v", neighbors)
	}

	if neighbors := dg.Successors(concrete.GonumNode(0)); len(neighbors) != 0 || dg.IsSuccessor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Errorf("Removing edge didn't properly remove successor")
	}

	if neighbors := dg.Predecessors(concrete.GonumNode(0)); len(neighbors) != 1 || neighbors[0].ID() != 2 ||
		!dg.IsPredecessor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Errorf("Removing directed edge improperly removed predecessor")
	}

	if neighbors := dg.Neighbors(concrete.GonumNode(2)); len(neighbors) != 1 || neighbors[0].ID() != 0 ||
		!dg.IsNeighbor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Errorf("Removing a directed edge removed reciprocal edge, neighbors: %v", neighbors)
	}

	if neighbors := dg.Successors(concrete.GonumNode(2)); len(neighbors) != 1 || neighbors[0].ID() != 0 ||
		!dg.IsSuccessor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Errorf("Removing edge improperly removed successor")
	}

	if neighbors := dg.Predecessors(concrete.GonumNode(2)); len(neighbors) != 0 || dg.IsPredecessor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Errorf("Removing directed edge wrongly kept predecessor")
	}

	dg.SetEdgeCost(concrete.GonumNode(0), concrete.GonumNode(2), 2.0, true)
	// I figure we've torture tested Neighbors/Successors/Predecessors at this point
	// so we'll just use the bool functions now
	if !dg.IsSuccessor(concrete.GonumNode(0), concrete.GonumNode(2)) {
		t.Error("Adding directed edge didn't change successor back")
	} else if !dg.IsSuccessor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Error("Adding directed edge strangely removed reverse successor")
	} else if c1, c2 := dg.Cost(concrete.GonumNode(2), concrete.GonumNode(0)), dg.Cost(concrete.GonumNode(0), concrete.GonumNode(2)); math.Abs(c1-c2) < .000001 {
		t.Error("Adding directed edge affected cost in undirected manner")
	}

	dg.RemoveEdge(concrete.GonumNode(2), concrete.GonumNode(0), false)
	if dg.IsSuccessor(concrete.GonumNode(0), concrete.GonumNode(2)) || dg.IsSuccessor(concrete.GonumNode(2), concrete.GonumNode(0)) {
		t.Error("Removing undirected edge did no work properly")
	}
}

type nodeSorter []graph.Node

func (ns nodeSorter) Len() int {
	return len(ns)
}

func (ns nodeSorter) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns nodeSorter) Less(i, j int) bool {
	return ns[i].ID() < ns[j].ID()
}

func TestDenseLists(t *testing.T) {
	dg := concrete.NewDenseGraph(15, true)
	nodes := nodeSorter(dg.NodeList())

	if len(nodes) != 15 {
		t.Fatalf("Wrong number of nodes")
	}

	sort.Sort(nodes)

	for i, node := range dg.NodeList() {
		if i != node.ID() {
			t.Errorf("Node list doesn't return properly id'd nodes")
		}
	}

	edges := dg.DirectedEdgeList()
	if len(edges) != 15*15 {
		t.Errorf("Improper number of edges for passable dense graph")
	}

	dg.RemoveEdge(concrete.GonumNode(12), concrete.GonumNode(11), true)
	edges = dg.DirectedEdgeList()
	if len(edges) != (15*15)-1 {
		t.Errorf("Removing edge didn't affect edge listing properly")
	}
}

func TestCrunch(t *testing.T) {
	dg := concrete.NewDenseGraph(5, true)
	dg.Crunch()
	if len(dg.NodeList()) != 5 || len(dg.DirectedEdgeList()) != 5*5 {
		t.Errorf("Crunch did something")
	}
}
