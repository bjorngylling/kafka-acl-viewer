package visjs

type nodeShape string

var (
	ShapeCircle nodeShape = "circle"
	ShapeBox    nodeShape = "box"

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
	ID    int       `json:"id"`
	Label string    `json:"label"`
	Shape nodeShape `json:"shape"`
	Color color     `json:"color"`
}

type Edge struct {
	From   int    `json:"from"`
	To     int    `json:"to"`
	Arrows string `json:"arrows"`
	Dashes bool   `json:"dashes,omitempty"`
	Title  string `json:"title,omitempty"`
}

type UserOps struct {
	To   map[string]struct{}
	From map[string]struct{}
}

func CreateNetwork(topics []string, userOperations map[string]UserOps) ([]Node, []Edge) {
	// Create user nodes, i.e. consumers and producers
	var nodes []Node
	i := 0
	userIdLookup := map[string]int{}
	for user := range userOperations {
		userIdLookup[user] = i
		nodes = append(nodes, Node{
			ID:    i,
			Label: "🤖 " + user,
			Shape: ShapeBox,
			Color: ColorGreen,
		})
		i++
	}

	// Create topic nodes
	topicIdLookup := map[string]int{}
	for _, topic := range topics {
		topicIdLookup[topic] = i
		nodes = append(nodes, Node{
			ID:    i,
			Label: "🗒 " + topic,
			Shape: ShapeBox,
		})
		i++
	}

	// Add all the edges
	var edges []Edge
	for user, ops := range userOperations {
		for input := range ops.From {
			edges = append(edges, Edge{
				From:   topicIdLookup[input],
				To:     userIdLookup[user],
				Arrows: "to",
				Dashes: false,
				Title:  "Read",
			})
		}
		for output := range ops.To {
			edges = append(edges, Edge{
				From:   userIdLookup[user],
				To:     topicIdLookup[output],
				Arrows: "to",
				Dashes: false,
				Title:  "Write",
			})
		}
	}

	return nodes, edges
}
