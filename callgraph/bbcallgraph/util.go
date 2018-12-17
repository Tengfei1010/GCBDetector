package bbcallgraph

import "github.com/Tengfei1010/GCBDetector/ssa"


func BBCallGraph(f *ssa.Function) *BBGraph {

	bg := New(nil)

	for _, bb := range f.Blocks {
		curNode := bg.CreateBBNode(bb)

		for _, suc := range bb.Succs {

			newBBNode := bg.CreateBBNode(suc)
			AddEdge(curNode, newBBNode)

		}
	}

	return bg
}


// GraphVisitEdges visits all the edges in BBGraph g in depth-first order.
// The edge function is called for each edge in postorder.  If it
// returns non-nil, visitation stops and GraphVisitEdges returns that
// value.
//
func GraphVisitEdges(g *BBGraph, edge func(*Edge) error) error {
	seen := make(map[*BBNode]bool)
	var visit func(n *BBNode) error
	visit = func(n *BBNode) error {
		if !seen[n] {
			seen[n] = true
			for _, e := range n.Out {
				if err := visit(e.Callee); err != nil {
					return err
				}
				if err := edge(e); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, n := range g.Nodes {
		if err := visit(n); err != nil {
			return err
		}
	}
	return nil
}

// PathSearch finds an arbitrary path starting at BBNode start and
// ending at some BBNode for which isEnd() returns true.  On success,
// PathSearch returns the path as an ordered list of edges; on
// failure, it returns nil.
//
func PathSearch(start *BBNode, isEnd func(*BBNode) bool) []*Edge {
	stack := make([]*Edge, 0, 32)
	seen := make(map[*BBNode]bool)
	var search func(n *BBNode) []*Edge
	search = func(n *BBNode) []*Edge {
		if !seen[n] {
			seen[n] = true
			if isEnd(n) {
				return stack
			}
			for _, e := range n.Out {
				stack = append(stack, e) // push
				if found := search(e.Callee); found != nil {
					return found
				}
				stack = stack[:len(stack)-1] // pop
			}
		}
		return nil
	}
	return search(start)
}

/*
  This function is used to search lock to lock path
 */
func LockPathSearch(start *BBNode, end *BBNode, lockKey string, filter func(*BBNode) bool) []*Edge {
	stack := make([]*Edge, 0, 32)
	seen := make(map[*BBNode]bool)
	var search func(n *BBNode) []*Edge
	search = func(n *BBNode) []*Edge {
		if !seen[n] {
			seen[n] = true
			if n == end {
				return stack
			}
			for _, e := range n.Out {
				if !filter(e.Callee) {
					continue
				}
				stack = append(stack, e) // push
				if found := search(e.Callee); found != nil {
					return found
				}
				stack = stack[:len(stack)-1] // pop
			}
		}
		return nil
	}
	return search(start)
}
