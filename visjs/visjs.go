package visjs

import (
	"github.com/bjorngylling/kafka-acl-viewer/graph"
)

type nodeShape string

var (
	ShapeBox nodeShape = "box"

	ColorGreen = color{Background: "#6ef091", Highlight: highlight{Background: "#ccffda"}}
)

type highlight struct {
	Background string `json:"background"`
}
type color struct {
	Background string    `json:"background"`
	Highlight  highlight `json:"highlight"`
}

type Node struct {
	ID    string    `json:"id"`
	Label string    `json:"label"`
	Shape nodeShape `json:"shape"`
	Color color     `json:"color"`
	Type  string    `json:"type"`
}

type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Arrows string `json:"arrows"`
	Dashes bool   `json:"dashes,omitempty"`
	Title  string `json:"title,omitempty"`
}

func CreateNetwork(graph graph.Graph) ([]Node, []Edge) {
	var nodes []Node
	var edges []Edge

	for id, node := range graph.Nodes {
		labelPrefix := ""
		color := color{}
		switch node.Type {
		case "user":
			labelPrefix = "🤖 "
			color = ColorGreen
		case "topic":
			labelPrefix = "🗒 "
		case "cluster":
			labelPrefix = "🗄 "
		}
		nodes = append(nodes, Node{
			ID:    id,
			Label: labelPrefix + node.Name,
			Shape: ShapeBox,
			Color: color,
			Type:  node.Type,
		})

		for _, edge := range node.Edges {
			edges = append(edges, Edge{
				From:   id,
				To:     edge.Target,
				Arrows: "to",
				Dashes: false,
				Title:  edge.Operation,
			})
		}
	}

	return nodes, edges
}
