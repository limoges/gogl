package gogl

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/fatih/set"
	. "launchpad.net/gocheck"
)

var _ = fmt.Println

// Hook gocheck into the go test runner
func Test(t *testing.T) { TestingT(t) }

var edgeSet = []Edge{
	BaseEdge{"foo", "bar"},
	BaseEdge{"bar", "baz"},
}

// swap method is useful for some testing shorthand
func (e BaseEdge) swap() Edge {
	return BaseEdge{e.V, e.U}
}

// This function automatically sets up suites of black box unit tests for
// graphs by determining which gogl interfaces they implement.
//
// Passing a graph to this method for testing is the most official way to
// determine whether or not it complies with not just the interfaces, but also
// the graph semantics defined by gogl.
func SetUpSimpleGraphTests(g Graph) bool {
	gf := &GraphFactory{g}
	var directed bool

	if dg, ok := g.(DirectedGraph); ok {
		directed = true
		Suite(&DirectedGraphSuite{Graph: dg, Factory: gf})
	}

	// Set up the basic Graph suite unconditionally
	Suite(&GraphSuite{Graph: g, Factory: gf, Directed: directed})

	if mg, ok := g.(MutableGraph); ok {
		Suite(&MutableGraphSuite{Graph: mg, Factory: gf, Directed: directed})
	}

	if wg, ok := g.(WeightedGraph); ok {
		Suite(&WeightedGraphSuite{Graph: wg, Factory: gf, Directed: directed})
	}

	if mwg, ok := g.(MutableWeightedGraph); ok {
		Suite(&MutableWeightedGraphSuite{Graph: mwg, Factory: gf, Directed: directed})
	}

	if sg, ok := g.(SimpleGraph); ok {
		Suite(&SimpleGraphSuite{Graph: sg, Factory: gf, Directed: directed})
	}

	return true
}

// Set up suites for all of gogl's graphs.
var _ = SetUpSimpleGraphTests(NewDirected())
var _ = SetUpSimpleGraphTests(NewUndirected())
var _ = SetUpSimpleGraphTests(NewWeightedDirected())
var _ = SetUpSimpleGraphTests(NewWeightedUndirected())

/* The GraphFactory - this generates graph instances for the tests. */

type GraphFactory struct {
	sourceGraph Graph
}

func (gf *GraphFactory) create() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(gf.sourceGraph)).Type()).Interface()
}

func (gf *GraphFactory) graphFromEdges(edges ...Edge) Graph {
	// For now just cheat and work through a Mutable interface
	base := gf.create()

	if mg, ok := base.(MutableGraph); ok {
		mg.AddEdges(edges...)
	} else if mwg, ok := base.(MutableWeightedGraph); ok {
		weighted_edges := make([]WeightedEdge, 0, len(edges))
		for _, edge := range edges {
			weighted_edges = append(weighted_edges, BaseWeightedEdge{BaseEdge{edge.Source(), edge.Target()}, 0})
		}
		mwg.AddEdges(weighted_edges...)
	} else {
		panic("Until GraphInitializers are made to work properly, all graphs have to be mutable for this testing harness to work.")
	}

	return base.(Graph)

}

func (gf *GraphFactory) CreateEmptyGraph() Graph {
	return gf.create().(Graph)
}

func (gf *GraphFactory) CreateGraphFromEdges(edges ...Edge) Graph {
	return gf.graphFromEdges(edges...)
}

func (gf *GraphFactory) CreateDirectedGraphFromEdges(edges ...Edge) DirectedGraph {
	return gf.graphFromEdges(edges...).(DirectedGraph)
}

func (gf *GraphFactory) CreateEmptySimpleGraph() SimpleGraph {
	return gf.create().(SimpleGraph)
}

func (gf *GraphFactory) CreateSimpleGraphFromEdges(edges ...Edge) SimpleGraph {
	return gf.graphFromEdges(edges...).(SimpleGraph)
}

func (gf *GraphFactory) CreateMutableGraph() MutableGraph {
	return gf.create().(MutableGraph)
}

func (gf *GraphFactory) CreateWeightedGraphFromEdges(edges ...WeightedEdge) WeightedGraph {
	base := gf.create()
	if mwg, ok := base.(MutableWeightedGraph); ok {
		mwg.AddEdges(edges...)
		return mwg
	}

	panic("Until GraphInitializers are made to work properly, all graphs have to be mutable for this testing harness to work.")
}

func (gf *GraphFactory) CreateEmptyWeightedGraph() WeightedGraph {
	return gf.create().(WeightedGraph)
}

func (gf *GraphFactory) CreateMutableWeightedGraph() MutableWeightedGraph {
	return gf.create().(MutableWeightedGraph)
}

/* Factory interfaces for tests */

type GraphCreator interface {
	CreateEmptyGraph() Graph
	CreateGraphFromEdges(edges ...Edge) Graph
}

type SimpleGraphCreator interface {
	CreateEmptySimpleGraph() SimpleGraph
	CreateSimpleGraphFromEdges(edges ...Edge) SimpleGraph
}

type DirectedGraphCreator interface {
	CreateDirectedGraphFromEdges(edges ...Edge) DirectedGraph
}

type MutableGraphCreator interface {
	CreateMutableGraph() MutableGraph
}

type WeightedGraphCreator interface {
	CreateEmptyWeightedGraph() WeightedGraph
	CreateWeightedGraphFromEdges(edges ...WeightedEdge) WeightedGraph
}

type MutableWeightedGraphCreator interface {
	CreateMutableWeightedGraph() MutableWeightedGraph
}

/* GraphSuite - tests for non-mutable graph methods */

type GraphSuite struct {
	Graph    Graph
	Factory  GraphCreator
	Directed bool
}

func (s *GraphSuite) SetUpTest(c *C) {
	s.Graph = s.Factory.CreateGraphFromEdges(edgeSet...)
}

func (s *GraphSuite) TestHasVertex(c *C) {
	c.Assert(s.Graph.HasVertex("qux"), Equals, false)
	c.Assert(s.Graph.HasVertex("foo"), Equals, true)
}

func (s *GraphSuite) TestHasEdge(c *C) {
	c.Assert(s.Graph.HasEdge(edgeSet[0]), Equals, true)
	c.Assert(s.Graph.HasEdge(BaseEdge{"qux", "quark"}), Equals, false)
}

func (s *GraphSuite) TestEachVertex(c *C) {
	var hit int
	f := func(v Vertex) {
		hit++
		c.Log("EachVertex hit closure, hit count", hit)
	}

	s.Graph.EachVertex(f)
	if !c.Check(hit, Equals, 3) {
		c.Error("EachVertex should have called injected closure iterator 3 times, actual count was ", hit)
	}
}

func (s *GraphSuite) TestEachEdge(c *C) {
	var hit int
	f := func(e Edge) {
		hit++
		c.Log("EachAdjacent hit closure with edge pair ", e.Source(), " ", e.Target(), " at hit count ", hit)
	}

	s.Graph.EachEdge(f)
	if !c.Check(hit, Equals, 2) {
		c.Error("EachEdge should have called injected closure iterator 2 times, actual count was ", hit)
	}
}

func (s *GraphSuite) TestEachAdjacent(c *C) {
	var hit int
	f := func(adj Vertex) {
		hit++
		c.Log("EachAdjacent hit closure with vertex ", adj, " at hit count ", hit)
	}

	s.Graph.EachAdjacent("foo", f)
	if !c.Check(hit, Equals, 1) {
		c.Error("EachEdge should have called injected closure iterator 2 times, actual count was ", hit)
	}
}

// This test is carefully constructed to be fully correct for directed graphs,
// and incidentally correct for undirected graphs.
func (s *GraphSuite) TestOutDegree(c *C) {
	g := s.Factory.CreateGraphFromEdges(&BaseEdge{"foo", "bar"})

	count, exists := g.OutDegree("foo")
	c.Assert(exists, Equals, true)
	c.Assert(count, Equals, 1)

	count, exists = g.OutDegree("missing")
	c.Assert(exists, Equals, false)
	c.Assert(count, Equals, 0)
}

// This test is carefully constructed to be fully correct for directed graphs,
// and incidentally correct for undirected graphs.
func (s *GraphSuite) TestInDegree(c *C) {
	g := s.Factory.CreateGraphFromEdges(&BaseEdge{"foo", "bar"})

	count, exists := g.InDegree("bar")
	c.Assert(exists, Equals, true)
	c.Assert(count, Equals, 1)

	count, exists = g.InDegree("missing")
	c.Assert(exists, Equals, false)
	c.Assert(count, Equals, 0)
}

func (s *GraphSuite) TestSize(c *C) {
	c.Assert(s.Graph.Size(), Equals, 2)

	g := s.Factory.CreateEmptyGraph()
	c.Assert(g.Size(), Equals, 0)
}

func (s *GraphSuite) TestOrder(c *C) {
	c.Assert(s.Graph.Size(), Equals, 2)

	g := s.Factory.CreateEmptyGraph()
	c.Assert(g.Size(), Equals, 0)
}

/* DirectedGraphSuite - tests for directed graph methods */

type DirectedGraphSuite struct {
	Graph   Graph
	Factory DirectedGraphCreator
}

func (s *DirectedGraphSuite) TestTranspose(c *C) {
	g := s.Factory.CreateDirectedGraphFromEdges(edgeSet...)

	g2 := g.Transpose()

	c.Assert(g2.HasEdge(edgeSet[0].(BaseEdge).swap()), Equals, true)
	c.Assert(g2.HasEdge(edgeSet[1].(BaseEdge).swap()), Equals, true)

	c.Assert(g2.HasEdge(edgeSet[0].(BaseEdge)), Equals, false)
	c.Assert(g2.HasEdge(edgeSet[1].(BaseEdge)), Equals, false)
}

/* SimpleGraphSuite - tests for simple graph methods */

type SimpleGraphSuite struct {
	Graph    Graph
	Factory  SimpleGraphCreator
	Directed bool
}

func (s *SimpleGraphSuite) TestDensity(c *C) {
	empty := s.Factory.CreateEmptySimpleGraph()
	c.Assert(math.IsNaN(empty.Density()), Equals, true)

	vev := s.Factory.CreateSimpleGraphFromEdges(BaseEdge{1, 2})
	if s.Directed {
		c.Assert(vev.Density(), Equals, float64(0.5))
	} else {
		c.Assert(vev.Density(), Equals, float64(1))
	}

	vevev := s.Factory.CreateSimpleGraphFromEdges(BaseEdge{1, 2}, BaseEdge{2, 3})
	if s.Directed {
		c.Assert(vevev.Density(), Equals, float64(2)/float64(6))
	} else {
		c.Assert(vevev.Density(), Equals, float64(2)/float64(3))
	}
}

/* MutableGraphSuite - tests for mutable graph methods */

type MutableGraphSuite struct {
	Graph    MutableGraph
	Factory  MutableGraphCreator
	Directed bool
}

func (s *MutableGraphSuite) TestEnsureVertex(c *C) {
	g := s.Factory.CreateMutableGraph()

	g.EnsureVertex("foo")
	c.Assert(g.HasVertex("foo"), Equals, true)
}

func (s *MutableGraphSuite) TestMultiEnsureVertex(c *C) {
	g := s.Factory.CreateMutableGraph()

	g.EnsureVertex("bar", "baz")
	c.Assert(g.HasVertex("bar"), Equals, true)
	c.Assert(g.HasVertex("baz"), Equals, true)
}

func (s *MutableGraphSuite) TestRemoveVertex(c *C) {
	g := s.Factory.CreateMutableGraph()

	g.EnsureVertex("bar", "baz")
	g.RemoveVertex("bar")
	c.Assert(g.HasVertex("bar"), Equals, false)
}

func (s *MutableGraphSuite) TestMultiRemoveVertex(c *C) {
	g := s.Factory.CreateMutableGraph()

	g.EnsureVertex("bar", "baz")
	g.RemoveVertex("bar", "baz")
	c.Assert(g.HasVertex("bar"), Equals, false)
	c.Assert(g.HasVertex("baz"), Equals, false)
}

func (s *MutableGraphSuite) TestAddAndRemoveEdge(c *C) {
	g := s.Factory.CreateMutableGraph()
	g.AddEdges(&BaseEdge{1, 2})

	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, true)
	c.Assert(g.HasEdge(BaseEdge{2, 1}), Equals, !s.Directed)

	// Now test removal
	g.RemoveEdges(&BaseEdge{1, 2})
	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, false)
	c.Assert(g.HasEdge(BaseEdge{2, 1}), Equals, false)
}

func (s *MutableGraphSuite) TestMultiAddAndRemoveEdge(c *C) {
	g := s.Factory.CreateMutableGraph()

	g.AddEdges(&BaseEdge{1, 2}, &BaseEdge{2, 3})

	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, true)
	c.Assert(g.HasEdge(BaseEdge{2, 3}), Equals, true)
	c.Assert(g.HasEdge(BaseEdge{2, 1}), Equals, !s.Directed)
	c.Assert(g.HasEdge(BaseEdge{3, 2}), Equals, !s.Directed)

	// Now test removal
	g.RemoveEdges(&BaseEdge{1, 2}, &BaseEdge{2, 3})
	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, false)
	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, false)
	c.Assert(g.HasEdge(BaseEdge{2, 3}), Equals, false)
	c.Assert(g.HasEdge(BaseEdge{2, 3}), Equals, false)
}

// Checks to ensure that removal works for both in-edges and out-edges.
func (s *MutableGraphSuite) TestVertexRemovalAlsoRemovesConnectedEdges(c *C) {
	g := s.Factory.CreateMutableGraph()

	g.AddEdges(&BaseEdge{1, 2}, &BaseEdge{2, 3}, &BaseEdge{4, 1})
	g.RemoveVertex(1)

	c.Assert(g.Size(), Equals, 1)
}

/* WeightedGraphSuite - tests for weighted graphs */

type WeightedGraphSuite struct {
	Graph    WeightedGraph
	Factory  WeightedGraphCreator
	Directed bool
}

func (s *WeightedGraphSuite) TestEachEdge(c *C) {
	// This method is not redundant with the base Graph suite as it ensures that the edges
	// provided by the EachEdge() iterator actually do implement WeightedEdge.
	g := s.Factory.CreateWeightedGraphFromEdges(BaseWeightedEdge{BaseEdge{1, 2}, 5}, BaseWeightedEdge{BaseEdge{2, 3}, -5})

	var we WeightedEdge
	g.EachEdge(func(e Edge) {
		c.Assert(e, Implements, &we)
	})
}

func (s *WeightedGraphSuite) TestEachWeightedEdge(c *C) {
	g := s.Factory.CreateWeightedGraphFromEdges(BaseWeightedEdge{BaseEdge{1, 2}, 5}, BaseWeightedEdge{BaseEdge{2, 3}, -5})

	edgeset := set.NewNonTS()
	g.EachWeightedEdge(func(e WeightedEdge) {
		edgeset.Add(e)
	})

	if s.Directed {
		c.Assert(edgeset.Has(BaseWeightedEdge{BaseEdge{1, 2}, 5}), Equals, true)
		c.Assert(edgeset.Has(BaseWeightedEdge{BaseEdge{2, 3}, -5}), Equals, true)
		c.Assert(edgeset.Has(BaseEdge{1, 2}), Equals, false)
		c.Assert(edgeset.Has(BaseEdge{2, 3}), Equals, false)
	} else {
		c.Assert(edgeset.Has(BaseWeightedEdge{BaseEdge{1, 2}, 5}) != edgeset.Has(BaseWeightedEdge{BaseEdge{2, 1}, 5}), Equals, true)
		c.Assert(edgeset.Has(BaseWeightedEdge{BaseEdge{2, 3}, -5}) != edgeset.Has(BaseWeightedEdge{BaseEdge{3, 2}, -5}), Equals, true)
		c.Assert(edgeset.Has(BaseEdge{1, 2}) || edgeset.Has(BaseEdge{2, 1}), Equals, false)
		c.Assert(edgeset.Has(BaseEdge{2, 3}) || edgeset.Has(BaseEdge{3, 2}), Equals, false)
	}
}

func (s *WeightedGraphSuite) TestHasWeightedEdge(c *C) {
	edges := []WeightedEdge{BaseWeightedEdge{BaseEdge{1, 2}, 5}}
	g := s.Factory.CreateWeightedGraphFromEdges(edges...)

	// TODO figure out how to meaningfully test undirected graphs' logic here
	c.Assert(g.HasWeightedEdge(edges[0]), Equals, true)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 1}), Equals, false) // wrong weight
}

/* MutableWeightedGraphSuite - tests for mutable weighted graphs */

type MutableWeightedGraphSuite struct {
	Graph    MutableWeightedGraph
	Factory  MutableWeightedGraphCreator
	Directed bool
}

func (s *MutableWeightedGraphSuite) TestEnsureVertex(c *C) {
	g := s.Factory.CreateMutableWeightedGraph()

	g.EnsureVertex("foo")
	c.Assert(g.HasVertex("foo"), Equals, true)
}

func (s *MutableWeightedGraphSuite) TestMultiEnsureVertex(c *C) {
	g := s.Factory.CreateMutableWeightedGraph()

	g.EnsureVertex("bar", "baz")
	c.Assert(g.HasVertex("bar"), Equals, true)
	c.Assert(g.HasVertex("baz"), Equals, true)
}

func (s *MutableWeightedGraphSuite) TestRemoveVertex(c *C) {
	g := s.Factory.CreateMutableWeightedGraph()

	g.EnsureVertex("bar", "baz")
	g.RemoveVertex("bar")
	c.Assert(g.HasVertex("bar"), Equals, false)
}

func (s *MutableWeightedGraphSuite) TestMultiRemoveVertex(c *C) {
	g := s.Factory.CreateMutableWeightedGraph()

	g.EnsureVertex("bar", "baz")
	g.RemoveVertex("bar", "baz")
	c.Assert(g.HasVertex("bar"), Equals, false)
	c.Assert(g.HasVertex("baz"), Equals, false)
}

func (s *MutableWeightedGraphSuite) TestAddAndRemoveEdge(c *C) {
	g := s.Factory.CreateMutableWeightedGraph()
	g.AddEdges(BaseWeightedEdge{BaseEdge{1, 2}, 5})

	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, true)
	c.Assert(g.HasEdge(BaseEdge{2, 1}), Equals, !s.Directed)

	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 5}), Equals, true)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 3}), Equals, false)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 1}, 5}), Equals, !s.Directed)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 1}, -3}), Equals, false)

	// Now test removal
	g.RemoveEdges(BaseWeightedEdge{BaseEdge{1, 2}, 5})
	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, false)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 5}), Equals, false)
}

func (s *MutableWeightedGraphSuite) TestMultiAddAndRemoveEdge(c *C) {
	g := s.Factory.CreateMutableWeightedGraph()
	g.AddEdges(BaseWeightedEdge{BaseEdge{1, 2}, 5}, BaseWeightedEdge{BaseEdge{2, 3}, -5})

	// Basic edge tests first
	// We test both Has*Edge() methods to ensure that adding our known edge fixture type results in the expected behavior.
	// Thus, this is not just duplicate testing of the Has*Edge() method.
	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, true)
	c.Assert(g.HasEdge(BaseEdge{2, 3}), Equals, true)
	c.Assert(g.HasEdge(BaseEdge{2, 1}), Equals, !s.Directed) // only if undirected
	c.Assert(g.HasEdge(BaseEdge{3, 2}), Equals, !s.Directed) // only if undirected

	// Now weighted edge tests
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 5}), Equals, true)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 3}), Equals, false) // wrong weight
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 1}, 5}), Equals, !s.Directed)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 1}, 3}), Equals, false) // wrong weight
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 3}, -5}), Equals, true)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 3}, 1}), Equals, false) // wrong weight
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{3, 2}, -5}), Equals, !s.Directed)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{3, 2}, 1}), Equals, false) // wrong weight

	// Now test removal
	g.RemoveEdges(BaseWeightedEdge{BaseEdge{1, 2}, 5}, BaseWeightedEdge{BaseEdge{2, 3}, -5})
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{1, 2}, 5}), Equals, false)
	c.Assert(g.HasWeightedEdge(BaseWeightedEdge{BaseEdge{2, 3}, -5}), Equals, false)
	c.Assert(g.HasEdge(BaseEdge{1, 2}), Equals, false)
	c.Assert(g.HasEdge(BaseEdge{2, 3}), Equals, false)
}
