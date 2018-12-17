/*

Package callgraph defines the call BBGraph and various algorithms
and utilities to operate on it.

A call BBGraph is a labelled directed BBGraph whose nodes represent
functions and whose edge labels represent syntactic function call
sites.  The presence of a labelled edge (caller, site, callee)
indicates that caller may call callee at the specified call site.

A call BBGraph is a multigraph: it may contain multiple edges (caller,
*, callee) connecting the same pair of nodes, so long as the edges
differ by label; this occurs when one function calls another function
from multiple call sites.  Also, it may contain multiple edges
(caller, site, *) that differ only by callee; this indicates a
polymorphic call.

A SOUND call BBGraph is one that overapproximates the dynamic calling
behaviors of the program in all possible executions.  One call BBGraph
is more PRECISE than another if it is a smaller overapproximation of
the dynamic behavior.

All call graphs have a synthetic root BBNode which is responsible for
calling main() and init().

Calls to built-in functions (e.g. panic, println) are not represented
in the call BBGraph; they are treated like built-in operators of the
language.

*/

/*
================================================================
=  Source code from https://github.com/dominikh/go-tools       =
=  Copyright @ Dominik Honnef (https://github.com/dominikh)    =
================================================================
*/

package bbcallgraph

// TODO(adonovan): add a function to eliminate wrappers from the
// callgraph, preserving topology.
// More generally, we could eliminate "uninteresting" nodes such as
// nodes from packages we don't care about.

import (
	"fmt"
	"go/token"

	"github.com/Tengfei1010/GCBDetector/ssa"
)

// A BBGraph represents a call BBGraph.
//
// A BBGraph may contain nodes that are not reachable from the root.
// If the call BBGraph is sound, such nodes indicate unreachable
// functions.
//
type BBGraph struct {
	Root  *BBNode                   // the distinguished root BBNode
	Nodes map[*ssa.BasicBlock]*BBNode // all nodes by function
}

// New returns a new BBGraph with the specified root BBNode.
func New(root *ssa.BasicBlock) *BBGraph {
	g := &BBGraph{Nodes: make(map[*ssa.BasicBlock]*BBNode)}
	g.Root = g.CreateBBNode(root)
	return g
}

// CreateNode returns the BBNode for fn, creating it if not present.
func (g *BBGraph) CreateBBNode(bb *ssa.BasicBlock) *BBNode {
	n, ok := g.Nodes[bb]
	if !ok {
		n = &BBNode{BB: bb, ID: len(g.Nodes)}
		g.Nodes[bb] = n
	}
	return n
}

// A BBNode represents a BBNode in a call BBGraph.
type BBNode struct {
	BB *ssa.BasicBlock // the basic block this BBNode represents
	ID   int           // 0-based sequence number
	In   []*Edge       // unordered set of incoming call edges (n.In[*].Callee == n)
	Out  []*Edge       // unordered set of outgoing call edges (n.Out[*].Caller == n)
}

func (n *BBNode) String() string {
	return fmt.Sprintf("n%d:%s", n.ID, n.BB)
}

// A Edge represents an edge in the call BBGraph.
//
// Site is nil for edges originating in synthetic or intrinsic
// functions, e.g. reflect.Call or the root of the call BBGraph.
type Edge struct {
	Caller *BBNode
	Callee *BBNode
}

func (e Edge) String() string {
	return fmt.Sprintf("%s --> %s", e.Caller, e.Callee)
}

func (e Edge) Description() string {
	return e.String()
}

func (e Edge) Pos() token.Pos {
	if e.Callee.BB.Instrs[0] == nil {
		return token.NoPos
	}
	site := e.Callee.BB.Instrs[0]
	return site.Pos()
}

// AddEdge adds the edge (caller, site, callee) to the call BBGraph.
// Elimination of duplicate edges is the caller's responsibility.
func AddEdge(caller *BBNode, callee *BBNode) {
	e := &Edge{caller, callee}
	callee.In = append(callee.In, e)
	caller.Out = append(caller.Out, e)
}
