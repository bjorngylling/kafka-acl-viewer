package graph

type Graph struct {
	Nodes map[string]*Node
}

type Node struct {
	Name  string
	Type  string
	Edges []*Edge
}

type Edge struct {
	Target    string
	Operation string
}

func NewGraph() Graph {
	return Graph{Nodes: map[string]*Node{}}
}

func (g *Graph) AddNode(n *Node) {
	g.Nodes[n.Name] = n
}

func (n *Node) AddEdge(to *Edge) {
	n.Edges = append(n.Edges, to)
}
