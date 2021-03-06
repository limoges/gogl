package gogl

import "fmt"

// for great justice
var fml = fmt.Println

/*
gogl's commonly used, high-level interfaces are composites of many smaller,
more atomic interfaces. Because there are so many atomic interfaces, we
start with the high-level composites. If you're new to gogl, you'll probably
want to read through the composite interfaces, jumping back and forth down
to the atomic interfaces as they appear in the composites. Patterns should
emerge quickly.

Note that not all atomic interfaces are used directly in the composites, so
make sure to review the atomics on their own, too.
*/

/* Composite graph interfaces */

// Graph is gogl's most basic interface: it contains only the methods that
// *every* type of graph implements.
//
// Graph is intentionally underspecified: both directed and undirected graphs
// implement it; simple, multi, and pseudographs; weighted, labeled, or any
// combination thereof.
//
// The semantics of some of these methods vary slightly from one graph type
// to another, but in general, the basic Graph methods are supplemented, not
// superceded, by the methods in more specific interfaces.
//
// Graph is a purely read oriented interface; the various Mutable*Graph
// interfaces contain the methods for writing.
type Graph interface {
	VertexEnumerator        // Enumerates vertices to an injected step function
	EdgeEnumerator          // Enumerates edges to an injected step function
	AdjacencyEnumerator     // Enumerates a vertex's adjacent vertices to an injected step function
	IncidentEdgeEnumerator  // Enumerates a vertex's incident edges to an injected step function
	VertexMembershipChecker // Allows inspection of contained vertices
	EdgeMembershipChecker   // Allows inspection of contained edges
	DegreeChecker           // Reports degree of vertices
}

// GraphSource is a subset of Graph, describing the minimal set of methods
// necessary to accomplish a naive full graph traversal and copy.
type GraphSource interface {
	VertexEnumerator
	EdgeEnumerator
}

// Digraph (directed graph) describes a Graph where all the edges are directed.
//
// gogl treats edge directionality as a property of the graph, not the edge itself.
// Thus, implementing this interface is gogl's only signal that a graph's edges are directed.
type Digraph interface {
	Graph
	ArcEnumerator         // Enumerates all arcs to an injected step function
	IncidentArcEnumerator // Enumerates a vertex's incident in- or out-arcs to an injected step function
	ProcessionEnumerator  // Enumerates a vertex's predecessor or successor vertices to a step function
	DirectedDegreeChecker // Reports in- and out-degree of vertices
	ArcMembershipChecker  // Allows inspection of contained arcs
	Transposer            // Digraphs can produce a transpose of themselves
}

// DigraphSource is a subset of Digraph, describing the minimal set of methods
// necessary to accomplish a naive full digraph traversal and copy.
type DigraphSource interface {
	GraphSource
	ArcEnumerator
}

// MutableGraph describes a graph with basic edges (no weighting, labeling, etc.)
// that can be modified freely by adding or removing vertices or edges.
type MutableGraph interface {
	Graph
	VertexSetMutator
	EdgeSetMutator
}

// MutableDigraph describes a digraph with basic arcs (no weighting, labeling, etc.)
// that can be modified freely by adding or removing vertices or arcs.
type MutableDigraph interface {
	Graph
	VertexSetMutator
	ArcSetMutator
}

// A simple graph is in opposition to a multigraph or pseudograph: it disallows loops and
// parallel edges.
type SimpleGraph interface {
	Graph
	Density() float64
}

// A weighted graph is a graph subtype where the edges have a numeric weight;
// as described by the WeightedEdge interface, this weight is a signed int.
//
// WeightedGraphs have both the HasEdge() and HasWeightedEdge() methods.
// Correct implementations should treat the difference as a matter of strictness:
//
// HasEdge() should return true as long as an edge exists
// connecting the two given vertices (respecting directed or undirected as
// appropriate), regardless of its weight.
//
// HasWeightedEdge() should return true iff an edge exists connecting the
// two given vertices (respecting directed or undirected as appropriate),
// AND if the edge weights are the same.
type WeightedGraph interface {
	Graph
	HasWeightedEdge(e WeightedEdge) bool
}

// WeightedDigraph describes a graph where all edges are weighted arcs (directed).
type WeightedDigraph interface {
	Digraph
	HasWeightedEdge(e WeightedEdge) bool
	HasWeightedArc(a WeightedArc) bool
}

// MutableWeightedGraph is the mutable version of a weighted graph. Its
// AddEdges() method is incompatible with MutableGraph, guaranteeing
// only weighted edges can be present in the graph.
type MutableWeightedGraph interface {
	WeightedGraph
	VertexSetMutator
	WeightedEdgeSetMutator
}

// A labeled graph is a graph subtype where the edges have an identifier;
// as described by the LabeledEdge interface, this identifier is a string.
//
// LabeledGraphs have both the HasEdge() and HasLabeledEdge() methods.
// Correct implementations should treat the difference as a matter of strictness:
//
// HasEdge() should return true as long as an edge exists
// connecting the two given vertices (respecting directed or undirected as
// appropriate), regardless of its label.
//
// HasLabeledEdge() should return true iff an edge exists connecting the
// two given vertices (respecting directed or undirected as appropriate),
// AND if the edge labels are the same.
type LabeledGraph interface {
	Graph
	HasLabeledEdge(e LabeledEdge) bool
}

// LabeledDigraph describes a graph where all edges are labeled arcs (directed).
type LabeledDigraph interface {
	Digraph
	HasLabeledEdge(e LabeledEdge) bool
	HasLabeledArc(a LabeledArc) bool
}

// MutableLabeledGraph is the mutable version of a labeled graph. Its
// AddEdges() method is incompatible with MutableGraph, guaranteeing
// only labeled edges can be present in the graph.
type MutableLabeledGraph interface {
	LabeledGraph
	VertexSetMutator
	LabeledEdgeSetMutator
}

// A data graph is a graph subtype where the edges carry arbitrary Go data;
// as described by the DataEdge interface, this identifier is an interface{}.
//
// DataGraphs have both the HasEdge() and HasDataEdge() methods.
// Correct implementations should treat the difference as a matter of strictness:
//
// HasEdge() should return true as long as an edge exists
// connecting the two given vertices (respecting directed or undirected as
// appropriate), regardless of its label.
//
// HasDataEdge() should return true iff an edge exists connecting the
// two given vertices (respecting directed or undirected as appropriate),
// AND if the edge data is the same. Simple comparison will typically be used
// to establish data equality, which means that using noncomparables (a slice,
// map, or non-pointer struct containing a slice or a map) for the data will
// cause a panic.
type DataGraph interface {
	Graph
	HasDataEdge(e DataEdge) bool
}

// DataDigraph describes a graph where all edges are data arcs (directed).
type DataDigraph interface {
	Digraph
	HasDataEdge(e DataEdge) bool
	HasDataArc(a DataArc) bool
}

// MutableDataGraph is the mutable version of a data graph. Its
// AddEdges() method is incompatible with MutableGraph, guaranteeing
// only property edges can be present in the graph.
type MutableDataGraph interface {
	DataGraph
	VertexSetMutator
	DataEdgeSetMutator
}

/* Atomic graph interfaces */

// EdgeSteps are used as arguments to various enumerators. They are called once for each edge produced by the enumerator.
//
// If the step function returns true, the calling enumerator is expected to end enumeration and return control to its caller.
type EdgeStep func(Edge) (terminate bool)

// ArcSteps are used as arguments to various enumerators. They are called once for each arc produced by the enumerator.
//
// If the step function returns true, the calling enumerator is expected to end enumeration and return control to its caller.
type ArcStep func(Arc) (terminate bool)

// VertexSteps are used as arguments to various enumerators. They are called once for each vertex produced by the enumerator.
//
// If the step function returns true, the calling enumerator is expected to end enumeration and return control to its caller.
type VertexStep func(Vertex) (terminate bool)

// A VertexEnumerator iteratively enumerates vertices.
type VertexEnumerator interface {
	// Calls the provided step function once with each vertex in the graph. Type
	// assert as appropriate in client code.
	Vertices(VertexStep)
}

// An EdgeEnumerator iteratively enumerates edges, and can indicate the number of edges present.
type EdgeEnumerator interface {
	// Calls the provided step function once with each edge in the graph. If a
	// specialized edge type (e.g., weighted) is known to be used by the
	// graph, it is the calling code's responsibility to type assert.
	Edges(EdgeStep)
}

// An ArcEnumerator iteratively enumerates edges, and can indicate the number of edges present.
type ArcEnumerator interface {
	// Calls the provided step function once with each edge in the graph. If a
	// specialized edge type (e.g., weighted) is known to be used by the
	// graph, it is the calling code's responsibility to type assert.
	Arcs(ArcStep)
}

// An IncidentEdgeEnumerator iteratively enumerates a given vertex's incident edges.
type IncidentEdgeEnumerator interface {
	// Calls the provided step function once with each edge incident to the
	// provided vertex. In a directed graph, this must include both
	// inbound and outbound edges.
	IncidentTo(v Vertex, incidentEdgeStep EdgeStep)
}

// An IncidentArcEnumerator iteratively enumerates a given vertex's incident arcs (directed edges).
// One enumerator provides inbound edges, the other outbound edges.
type IncidentArcEnumerator interface {
	// Calls the provided step function once with each arc outbound from the
	// provided vertex.
	ArcsFrom(v Vertex, outArcStep ArcStep)
	// Calls the provided step function once with each arc outbound from the
	// provided vertex.
	ArcsTo(v Vertex, inArcStep ArcStep)
}

// A ProcessionEnumerator iteratively enumerates a vertex's predecessors or successors
// into an injected step function.
type ProcessionEnumerator interface { // TODO ProcessionEnumerator? really?
	SuccessorsOf(v Vertex, successorStep VertexStep)
	PredecessorsOf(v Vertex, predecessorStep VertexStep)
}

// An AdjacencyEnumerator iteratively enumerates a given vertex's adjacent vertices.
type AdjacencyEnumerator interface {
	// Calls the provided step function once with each vertex adjacent to the
	// the provided vertex. In a digraph, this includes both successor
	// and predecessor vertices.
	AdjacentTo(start Vertex, adjacentVertexStep VertexStep)
}

// A VertexMembershipChecker can indicate the presence of a vertex.
type VertexMembershipChecker interface {
	// Indicates whether or not the vertex is present in the graph.
	HasVertex(Vertex) bool
}

// A DegreeChecker reports the number of edges incident to a given vertex.
type DegreeChecker interface {
	DegreeOf(Vertex) (degree int, exists bool) // Number of incident edges; if vertex is present
}

// A DirectedDegreeChecker reports the number of in or out-edges incident to given vertex.
type DirectedDegreeChecker interface {
	InDegreeOf(Vertex) (degree int, exists bool)  // Number of in-edges; if vertex is present
	OutDegreeOf(Vertex) (degree int, exists bool) // Number of out-edges; if vertex is present
}

// An EdgeMembershipChecker can indicate the presence of an edge.
type EdgeMembershipChecker interface {
	HasEdge(Edge) bool
}

// An ArcMembershipChecker can indicate the presence of an arc.
type ArcMembershipChecker interface {
	HasArc(Arc) bool
}

// A VertexSetMutator allows the addition and removal of vertices from a set.
type VertexSetMutator interface {
	// Ensures the provided vertices are present in the graph.
	EnsureVertex(...Vertex)
	// Removes the provided vertices from the graph, if present.
	RemoveVertex(...Vertex)
}

// An EdgeSetMutator allows the addition and removal of edges from a set.
type EdgeSetMutator interface {
	AddEdges(edges ...Edge)
	RemoveEdges(edges ...Edge)
}

// An ArcSetMutator allows the addition and removal of arcs from a set.
type ArcSetMutator interface {
	AddArcs(arcs ...Arc)
	RemoveArcs(arcs ...Arc)
}

// A WeightedEdgeSetMutator allows the addition and removal of weighted edges from a set.
type WeightedEdgeSetMutator interface {
	AddEdges(edges ...WeightedEdge)
	RemoveEdges(edges ...WeightedEdge)
}

// A WeightedArcSetMutator allows the addition and removal of weighted arcs from a set.
type WeightedArcSetMutator interface {
	AddArcs(arcs ...WeightedArc)
	RemoveArcs(arcs ...WeightedArc)
}

// A LabeledEdgeSetMutator allows the addition and removal of labeled edges from a set.
type LabeledEdgeSetMutator interface {
	AddEdges(edges ...LabeledEdge)
	RemoveEdges(edges ...LabeledEdge)
}

// A LabeledArcSetMutator allows the addition and removal of labeled arcs from a set.
type LabeledArcSetMutator interface {
	AddArcs(arcs ...LabeledArc)
	RemoveArcs(arcs ...LabeledArc)
}

// A DataEdgeSetMutator allows the addition and removal of data edges from a set.
type DataEdgeSetMutator interface {
	AddEdges(edges ...DataEdge)
	RemoveEdges(edges ...DataEdge)
}

// A DataArcSetMutator allows the addition and removal of data arcs from a set.
type DataArcSetMutator interface {
	AddArcs(arcs ...DataArc)
	RemoveArcs(arcs ...DataArc)
}

/* Optional optimization interfaces

These interfaces describe behaviors and information about a graph which can be
naively calculated/performed using the enumeration methods, but where a particular
implementation may be able to perform that operation more efficiently using tricks
specific to the underling graph implementation.

In other words, graph structures SHOULD implement any method in this set that they
can perform more efficiently than a full linear traversal.

gogl's general goal is to provide one or more standalone functors for each of the
capabilities described here. These functors are aware of the optional optimization
interfaces, and wlil automatically use them if available. Client code is expected
and encouraged to take care of these functors where possible.
*/

// A VertexCounter provides a numeric count of the number of unique vertices in a graph.
type VertexCounter interface {
	Order() int
}

// An EdgeCounter provides a numeric count of the number of unique edges in a graph.
type EdgeCounter interface {
	Size() int
}

// A Transposer produces a transposed version of a Digraph.
type Transposer interface {
	Transpose() Digraph
}
