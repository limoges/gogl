package gogl

import "sync"

type adjacencyList map[Vertex]VertexSet

// Helper to not have to write struct{} everywhere.
var keyExists = struct{}{}

type al struct {
	list adjacencyList
}

type AdjacencyList struct {
	al
	size uint
	mu   sync.RWMutex
}

// Composite literal to create a new AdjacencyList.
func NewAdjacencyList() *AdjacencyList {
	// Cannot assign to promoted fields in a composite literals.
	list := &AdjacencyList{}
	list.list = make(map[Vertex]VertexSet)
	return list
}

func NewAdjacencyListFromEdgeSet(set []Edge) *AdjacencyList {
	g := NewAdjacencyList()

	for _, edge := range set {
		g.addEdge(edge)
	}

	return g
}

func (g *AdjacencyList) EachVertex(f func(vertex Vertex)) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for v := range g.list {
		f(v)
	}
}

func (g *AdjacencyList) EachEdge(f func(edge Edge)) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for source, adjacent := range g.list {
		for target, _ := range adjacent {
			f(&BaseEdge{u: source, v: target})
		}
	}
}

func (g *AdjacencyList) EachAdjacent(vertex Vertex, f func(target Vertex)) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.list[vertex]; exists {
		for adjacent, _ := range g.list[vertex] {
			f(adjacent)
		}
	}
}

func (g *AdjacencyList) HasVertex(vertex Vertex) (exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	exists = g.hasVertex(vertex)
	return
}

func (g *AdjacencyList) OutDegree(vertex Vertex) (degree uint, exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if exists = g.hasVertex(vertex); exists {
		degree = uint(len(g.list[vertex]))
	}
	return
}

// Getting InDegree is inefficient for directed adjacency lists
func (g *AdjacencyList) InDegree(vertex Vertex) (degree uint, exists bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if exists = g.hasVertex(vertex); exists {

		f := func(v Vertex) {
			if v == vertex {
				degree++
			}
		}

		// This results in a double read-lock. Should be fine.
		for e := range g.list {
			g.EachAdjacent(e, f)
		}
	}

	return
}

func (g *AdjacencyList) hasVertex(vertex Vertex) (exists bool) {
	_, exists = g.list[vertex]
	return
}

func (g *AdjacencyList) Order() uint {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return uint(len(g.list))
}

func (g *AdjacencyList) Size() uint {
	return g.size
}

func (g *AdjacencyList) Density() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	order := g.Order()
	return 2 * float64(g.Size()) / float64(order * (order - 1))
}

func (g *AdjacencyList) AddVertex(vertex Vertex) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.addVertex(vertex)
}

func (g *AdjacencyList) addVertex(vertex Vertex) (success bool) {
	if exists := g.hasVertex(vertex); !exists {
		// TODO experiment with different lengths...possibly by analyzing existing density?
		g.list[vertex] = make(VertexSet, 10)
		success = true
	}

	return
}

func (g *AdjacencyList) RemoveVertex(vertex Vertex) (success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.hasVertex(vertex) {
		// TODO Is the expensive search good to do here and now...
		// while read-locked?
		delete(g.list, vertex)

		// TODO consider chunking the list and parallelizing into goroutines
		for _, adjacent := range g.list {
			if _, has := adjacent[vertex]; has {
				delete(adjacent, vertex)
				g.size--
			}
		}

		success = true
	}
	return
}

func (g *AdjacencyList) AddEdge(edge Edge) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.addEdge(edge)
}

func (g *AdjacencyList) addEdge(edge Edge) (exists bool) {
	g.addVertex(edge.Source())
	g.addVertex(edge.Target())

	if _, exists = g.list[edge.Source()][edge.Target()]; !exists {
		g.list[edge.Source()][edge.Target()] = keyExists
		g.size++
	}
	return !exists
}

func (g *AdjacencyList) RemoveEdge(edge Edge) {
	g.mu.Lock()
	defer g.mu.Unlock()

	s, t := edge.Both()
	if _, exists := g.list[s][t]; exists {
		delete(g.list[s], t)
		g.size--
	}
}

