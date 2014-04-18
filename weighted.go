package gogl

import (
	"sync"

	"gopkg.in/fatih/set.v0"
)

// This is implemented as an adjacency list, because those are simple.
type baseWeighted struct {
	list map[Vertex]map[Vertex]int
	size int
	mu   sync.RWMutex
}

/* baseWeighted shared methods */

// Traverses the graph's vertices in random order, passing each vertex to the
// provided closure.
func (g *baseWeighted) EachVertex(f VertexLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for v := range g.list {
		if f(v) {
			return
		}
	}
}

// Indicates whether or not the given vertex is present in the graph.
func (g *baseWeighted) HasVertex(vertex Vertex) (exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	exists = g.hasVertex(vertex)
	return
}

// Indicates whether or not the given vertex is present in the graph.
func (g *baseWeighted) hasVertex(vertex Vertex) (exists bool) {
	_, exists = g.list[vertex]
	return
}

// Returns the order (number of vertices) in the graph.
func (g *baseWeighted) Order() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.list)
}

// Returns the size (number of edges) in the graph.
func (g *baseWeighted) Size() int {
	return g.size
}

// Adds the provided vertices to the graph. If a provided vertex is
// already present in the graph, it is a no-op (for that vertex only).
func (g *baseWeighted) EnsureVertex(vertices ...Vertex) {
	if len(vertices) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.ensureVertex(vertices...)
}

// Adds the provided vertices to the graph. If a provided vertex is
// already present in the graph, it is a no-op (for that vertex only).
func (g *baseWeighted) ensureVertex(vertices ...Vertex) {
	// TODO this is horrible, but the reflection approach in the testing harness requires it...for now
	if g.list == nil {
		g.list = make(map[Vertex]map[Vertex]int)
	}

	for _, vertex := range vertices {
		if !g.hasVertex(vertex) {
			// TODO experiment with different lengths...possibly by analyzing existing density?
			g.list[vertex] = make(map[Vertex]int, 10)
		}
	}

	return
}

/* DirectedWeighted implementation */

type weightedDirected struct {
	baseWeighted
}

func NewWeightedDirected() MutableWeightedGraph {
	list := &weightedDirected{}
	// Cannot assign to promoted fields in a composite literals.
	list.list = make(map[Vertex]map[Vertex]int)

	// Type assertions to ensure interfaces are met
	var _ Graph = list
	var _ SimpleGraph = list
	var _ DirectedGraph = list
	var _ WeightedGraph = list
	var _ MutableWeightedGraph = list

	return list
}

// Returns the outdegree of the provided vertex. If the vertex is not present in the
// graph, the second return value will be false.
func (g *weightedDirected) OutDegreeOf(vertex Vertex) (degree int, exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if exists = g.hasVertex(vertex); exists {
		degree = len(g.list[vertex])
	}
	return
}

// Returns the indegree of the provided vertex. If the vertex is not present in the
// graph, the second return value will be false.
//
// Note that getting indegree is inefficient for directed adjacency lists; it requires
// a full scan of the graph's edge set.
func (g *weightedDirected) InDegreeOf(vertex Vertex) (degree int, exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if exists = g.hasVertex(vertex); exists {
		g.EachEdge(func(e Edge) (terminate bool) {
			if vertex == e.Target() {
				degree++
			}
			return
		})
	}

	return
}

// Returns the degree of the given vertex, counting both in and out-edges.
func (g *weightedDirected) DegreeOf(vertex Vertex) (degree int, exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	indegree, exists := g.InDegreeOf(vertex)
	outdegree, exists := g.OutDegreeOf(vertex)
	return indegree + outdegree, exists
}

// Enumerates the set of all edges incident to the provided vertex.
func (g *weightedDirected) EachEdgeIncidentTo(v Vertex, f EdgeLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var terminate bool
	interloper := func(e Edge) bool {
		terminate = terminate || f(e)
		return terminate
	}

	g.EachArcFrom(v, interloper)
	g.EachArcTo(v, interloper)
}

// Enumerates the vertices adjacent to the provided vertex.
func (g *weightedDirected) EachAdjacent(start Vertex, f VertexLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	g.EachEdgeIncidentTo(start, func(e Edge) bool {
		u, v := e.Both()
		if u == start {
			return f(v)
		} else {
			return f(u)
		}
	})
}

// Enumerates the set of out-edges for the provided vertex.
func (g *weightedDirected) EachArcFrom(v Vertex, f EdgeLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.hasVertex(v) {
		return
	}

	for adjacent, weight := range g.list[v] {
		if f(BaseWeightedEdge{BaseEdge{U: v, V: adjacent}, weight}) {
			return
		}
	}
}

// Enumerates the set of in-edges for the provided vertex.
func (g *weightedDirected) EachArcTo(v Vertex, f EdgeLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.hasVertex(v) {
		return
	}

	for candidate, adjacent := range g.list {
		for target, weight := range adjacent {
			if target == v {
				if f(BaseWeightedEdge{BaseEdge{U: candidate, V: target}, weight}) {
					return
				}
			}
		}
	}
}

// Traverses the set of edges in the graph, passing each edge to the
// provided closure.
func (g *weightedDirected) EachEdge(f EdgeLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for source, adjacent := range g.list {
		for target, weight := range adjacent {
			if f(BaseWeightedEdge{BaseEdge{U: source, V: target}, weight}) {
				return
			}
		}
	}
}

// Traverses the set of edges in the graph, passing each edge and its weight
// to the provided closure.
func (g *weightedDirected) EachWeightedEdge(f func(edge WeightedEdge)) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for source, adjacent := range g.list {
		for target, weight := range adjacent {
			f(BaseWeightedEdge{BaseEdge{U: source, V: target}, weight})
		}
	}
}

// Indicates whether or not the given edge is present in the graph. It matches
// based solely on the presence of an edge, disregarding edge weight.
func (g *weightedDirected) HasEdge(edge Edge) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	_, exists := g.list[edge.Source()][edge.Target()]
	return exists
}

// Indicates whether or not the given weighted edge is present in the graph.
// It will only match if the provided WeightedEdge has the same weight as
// the edge contained in the graph.
func (g *weightedDirected) HasWeightedEdge(edge WeightedEdge) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if weight, exists := g.list[edge.Source()][edge.Target()]; exists {
		return weight == edge.Weight()
	}
	return false
}

// Returns the density of the graph. Density is the ratio of edge count to the
// number of edges there would be in complete graph (maximum edge count).
func (g *weightedDirected) Density() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	order := g.Order()
	return float64(g.Size()) / float64(order*(order-1))
}

// Removes a vertex from the graph. Also removes any edges of which that
// vertex is a member.
func (g *weightedDirected) RemoveVertex(vertices ...Vertex) {
	if len(vertices) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for _, vertex := range vertices {
		if g.hasVertex(vertex) {
			g.size -= len(g.list[vertex])
			delete(g.list, vertex)

			for _, adjacent := range g.list {
				if _, has := adjacent[vertex]; has {
					delete(adjacent, vertex)
					g.size--
				}
			}
		}
	}
	return
}

// Adds edges to the graph.
func (g *weightedDirected) AddEdges(edges ...WeightedEdge) {
	if len(edges) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.addEdges(edges...)
}

// Adds a new edge to the graph.
func (g *weightedDirected) addEdges(edges ...WeightedEdge) {
	for _, edge := range edges {
		g.ensureVertex(edge.Source(), edge.Target())

		if _, exists := g.list[edge.Source()][edge.Target()]; !exists {
			g.list[edge.Source()][edge.Target()] = edge.Weight()
			g.size++
		}
	}
}

// Removes edges from the graph. This does NOT remove vertex members of the
// removed edges.
func (g *weightedDirected) RemoveEdges(edges ...WeightedEdge) {
	if len(edges) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for _, edge := range edges {
		s, t := edge.Both()
		if _, exists := g.list[s][t]; exists {
			delete(g.list[s], t)
			g.size--
		}
	}
}

func (g *weightedDirected) Transpose() DirectedGraph {
	g.mu.RLock()
	defer g.mu.RUnlock()

	g2 := &weightedDirected{}
	g2.list = make(map[Vertex]map[Vertex]int)

	// Guess at average indegree by looking at ratio of edges to vertices, use that to initially size the adjacency maps
	startcap := int(g.Size() / g.Order())

	for source, adjacent := range g.list {
		if !g2.hasVertex(source) {
			g2.list[source] = make(map[Vertex]int, startcap+1)
		}

		for target, weight := range adjacent {
			if !g2.hasVertex(target) {
				g2.list[target] = make(map[Vertex]int, startcap+1)
			}
			g2.list[target][source] = weight
		}
	}

	return g2
}

/* UndirectedWeighted implementation */

type weightedUndirected struct {
	baseWeighted
}

func NewWeightedUndirected() MutableWeightedGraph {
	g := &weightedUndirected{}
	// Cannot assign to promoted fields in a composite literals.
	g.list = make(map[Vertex]map[Vertex]int)

	// Type assertions to ensure interfaces are met
	var _ Graph = g
	var _ SimpleGraph = g
	var _ WeightedGraph = g
	var _ MutableWeightedGraph = g

	return g
}

// Returns the degree of the provided vertex. If the vertex is not present in the
// graph, the second return value will be false.
func (g *weightedUndirected) DegreeOf(vertex Vertex) (degree int, exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if exists = g.hasVertex(vertex); exists {
		degree = len(g.list[vertex])
	}
	return
}

// Traverses the set of edges in the graph, passing each edge to the
// provided closure.
func (g *weightedUndirected) EachEdge(f EdgeLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := set.NewNonTS()

	for source, adjacent := range g.list {
		for target, weight := range adjacent {
			be := BaseEdge{U: source, V: target}
			e := BaseWeightedEdge{be, weight}
			if !visited.Has(BaseEdge{U: target, V: source}) {
				visited.Add(be)
				if f(e) {
					return
				}
			}
		}
	}
}

// Traverses the set of edges in the graph, passing each edge and its weight
// to the provided closure.
func (g *weightedUndirected) EachWeightedEdge(f func(edge WeightedEdge)) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := set.NewNonTS()

	for source, adjacent := range g.list {
		for target, weight := range adjacent {
			be := BaseEdge{U: source, V: target}
			e := BaseWeightedEdge{be, weight}
			if !visited.Has(BaseEdge{U: target, V: source}) {
				visited.Add(be)
				f(e)
			}
		}
	}
}

// Enumerates the set of all edges incident to the provided vertex.
func (g *weightedUndirected) EachEdgeIncidentTo(v Vertex, f EdgeLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.hasVertex(v) {
		return
	}

	for adjacent, weight := range g.list[v] {
		if f(BaseWeightedEdge{BaseEdge{U: v, V: adjacent}, weight}) {
			return
		}
	}
}

// Enumerates the vertices adjacent to the provided vertex.
func (g *weightedUndirected) EachAdjacent(vertex Vertex, f VertexLambda) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	g.eachAdjacent(vertex, f)
}

func (g *weightedUndirected) eachAdjacent(vertex Vertex, f VertexLambda) {
	if _, exists := g.list[vertex]; exists {
		for adjacent, _ := range g.list[vertex] {
			if f(adjacent) {
				return
			}
		}
	}
}

// Indicates whether or not the given edge is present in the graph. It matches
// based solely on the presence of an edge, disregarding edge weight.
func (g *weightedUndirected) HasEdge(edge Edge) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Spread it into two expressions to avoid evaluating the second if possible
	if _, exists := g.list[edge.Source()][edge.Target()]; exists {
		return true
	} else if _, exists := g.list[edge.Target()][edge.Source()]; exists {
		return true
	}
	return false
}

// Indicates whether or not the given weighted edge is present in the graph.
// It will only match if the provided WeightedEdge has the same weight as
// the edge contained in the graph.
func (g *weightedUndirected) HasWeightedEdge(edge WeightedEdge) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Spread it into two expressions to avoid evaluating the second if possible
	if weight, exists := g.list[edge.Source()][edge.Target()]; exists {
		return edge.Weight() == weight
	} else if weight, exists := g.list[edge.Target()][edge.Source()]; exists {
		return edge.Weight() == weight
	}
	return false
}

// Returns the density of the graph. Density is the ratio of edge count to the
// number of edges there would be in complete graph (maximum edge count).
func (g *weightedUndirected) Density() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	order := g.Order()
	return 2 * float64(g.Size()) / float64(order*(order-1))
}

// Removes a vertex from the graph. Also removes any edges of which that
// vertex is a member.
func (g *weightedUndirected) RemoveVertex(vertices ...Vertex) {
	if len(vertices) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for _, vertex := range vertices {
		if g.hasVertex(vertex) {
			g.eachAdjacent(vertex, func(adjacent Vertex) (terminate bool) {
				delete(g.list[adjacent], vertex)
				return
			})
			g.size -= len(g.list[vertex])
			delete(g.list, vertex)
		}
	}
	return
}

// Adds edges to the graph.
func (g *weightedUndirected) AddEdges(edges ...WeightedEdge) {
	if len(edges) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.addEdges(edges...)
}

// Adds a new edge to the graph.
func (g *weightedUndirected) addEdges(edges ...WeightedEdge) {
	for _, edge := range edges {
		g.ensureVertex(edge.Source(), edge.Target())

		if _, exists := g.list[edge.Source()][edge.Target()]; !exists {
			w := edge.Weight()
			g.list[edge.Source()][edge.Target()] = w
			g.list[edge.Target()][edge.Source()] = w
			g.size++
		}
	}
}

// Removes edges from the graph. This does NOT remove vertex members of the
// removed edges.
func (g *weightedUndirected) RemoveEdges(edges ...WeightedEdge) {
	if len(edges) == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for _, edge := range edges {
		s, t := edge.Both()
		if _, exists := g.list[s][t]; exists {
			delete(g.list[s], t)
			delete(g.list[t], s)
			g.size--
		}
	}
}
