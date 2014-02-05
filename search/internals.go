package search

import (
	"container/heap"

	gr "github.com/gonum/graph"
)

// Sets up the cost functions and successor functions so I don't have to do a type switch every time.
// This almost always does more work than is necessary, but since it's only executed once per function, and graph functions are rather costly, the "extra work"
// should be negligible.
func setupFuncs(graph gr.Graph, cost, heuristicCost func(gr.Node, gr.Node) float64) (successorsFunc, predecessorsFunc, neighborsFunc func(gr.Node) []gr.Node, isSuccessorFunc, isPredecessorFunc,
	isNeighborFunc func(gr.Node, gr.Node) bool,
	costFunc, heuristicCostFunc func(gr.Node, gr.Node) float64) {

	switch g := graph.(type) {
	case gr.DirectedGraph:
		successorsFunc = g.Successors
		predecessorsFunc = g.Predecessors
		neighborsFunc = g.Neighbors
		isSuccessorFunc = g.IsSuccessor
		isPredecessorFunc = g.IsPredecessor
		isNeighborFunc = g.IsNeighbor
	default:
		successorsFunc = g.Neighbors
		predecessorsFunc = g.Neighbors
		neighborsFunc = g.Neighbors
		isSuccessorFunc = g.IsNeighbor
		isPredecessorFunc = g.IsNeighbor
		isNeighborFunc = g.IsNeighbor
	}

	if heuristicCost != nil {
		heuristicCostFunc = heuristicCost
	} else {
		if g, ok := graph.(gr.HeuristicCoster); ok {
			heuristicCostFunc = g.HeuristicCost
		} else {
			heuristicCostFunc = NullHeuristic
		}
	}

	if cost != nil {
		costFunc = cost
	} else {
		if g, ok := graph.(gr.Coster); ok {
			costFunc = g.Cost
		} else {
			costFunc = UniformCost
		}
	}

	return
}

/* Purely internal data structures and functions (mostly for sorting) */

// A package that contains an edge (as from EdgeList), and a Weight (as if Cost(Edge.Head(), Edge.Tail()) had been called)
type WeightedEdge struct {
	gr.Edge
	Weight float64
}

/** Sorts a list of edges by weight, agnostic to repeated edges as well as direction **/

type edgeSorter []WeightedEdge

func (el edgeSorter) Len() int {
	return len(el)
}

func (el edgeSorter) Less(i, j int) bool {
	return el[i].Weight < el[j].Weight
}

func (el edgeSorter) Swap(i, j int) {
	el[i], el[j] = el[j], el[i]
}

/** Keeps track of a node's scores so they can be used in a priority queue for A* **/

type internalNode struct {
	gr.Node
	gscore, fscore float64
}

/* A* stuff */
type aStarPriorityQueue struct {
	indexList map[int]int
	nodes     []internalNode
}

func (pq *aStarPriorityQueue) Less(i, j int) bool {
	return pq.nodes[i].fscore < pq.nodes[j].fscore // As the heap documentation says, a priority queue is listed if the actual values are treated as if they were negative
}

func (pq *aStarPriorityQueue) Swap(i, j int) {
	pq.indexList[pq.nodes[i].ID()] = j
	pq.indexList[pq.nodes[j].ID()] = i

	pq.nodes[i], pq.nodes[j] = pq.nodes[j], pq.nodes[i]
}

func (pq *aStarPriorityQueue) Len() int {
	return len(pq.nodes)
}

func (pq *aStarPriorityQueue) Push(x interface{}) {
	node := x.(internalNode)
	pq.nodes = append(pq.nodes, node)
	pq.indexList[node.ID()] = len(pq.nodes) - 1
}

func (pq *aStarPriorityQueue) Pop() interface{} {
	x := pq.nodes[len(pq.nodes)-1]
	pq.nodes = pq.nodes[:len(pq.nodes)-1]
	delete(pq.indexList, x.ID())

	return x
}

func (pq *aStarPriorityQueue) Fix(id int, newGScore, newFScore float64) {
	if i, ok := pq.indexList[id]; ok {
		pq.nodes[i].gscore = newGScore
		pq.nodes[i].fscore = newFScore
		heap.Fix(pq, i)
	}
}

func (pq *aStarPriorityQueue) Find(id int) (internalNode, bool) {
	loc, ok := pq.indexList[id]
	if ok {
		return pq.nodes[loc], true
	} else {
		return internalNode{}, false
	}

}

func (pq *aStarPriorityQueue) Exists(id int) bool {
	_, ok := pq.indexList[id]
	return ok
}

// General utility funcs

// Rebuilds a path backwards from the goal.
func rebuildPath(predecessors map[int]gr.Node, goal gr.Node) []gr.Node {
	if n, ok := goal.(internalNode); ok {
		goal = n.Node
	}
	path := []gr.Node{goal}
	curr := goal
	for prev, ok := predecessors[curr.ID()]; ok; prev, ok = predecessors[curr.ID()] {
		if n, ok := prev.(internalNode); ok {
			prev = n.Node
		}
		path = append(path, prev)
		curr = prev
	}

	// Reverse the path since it was built backwards
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}
