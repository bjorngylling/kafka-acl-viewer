package visjs

import (
	"github.com/bjorngylling/kafka-acl-viewer/graph"
	"reflect"
	"testing"
)

func Test_createNetwork(t *testing.T) {
	type args struct {
		graph graph.Graph
	}

	tests := []struct {
		name      string
		args      args
		wantNodes []Node
		wantEdges []Edge
	}{
		{
			name: "basic",
			args: args{
				graph: graph.Graph{
					Nodes: map[string]*graph.Node{
						"user-1": {
							Name: "user-1",
							Type: "user",
						},
						"topic-1": {
							Name:  "topic-1",
							Edges: []*graph.Edge{{Target: "user-1", Operation: "Read"}},
							Type:  "topic",
						},
					},
				},
			},
			wantNodes: []Node{
				{
					ID:    "user-1",
					Label: "🤖 user-1",
					Shape: "box",
					Color: color{"#6ef091", highlight{Background: "#ccffda"}},
				},
				{
					ID:    "topic-1",
					Label: "🗒 topic-1",
					Shape: "box",
					Color: color{"", highlight{""}},
				},
			},
			wantEdges: []Edge{
				{
					From:   "topic-1",
					To:     "user-1",
					Arrows: "to",
					Dashes: false,
					Title:  "Read",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CreateNetwork(tt.args.graph)
			if !reflect.DeepEqual(got, tt.wantNodes) {
				t.Errorf("createNetwork() got = %v, wantNodes %v", got, tt.wantNodes)
			}
			if !reflect.DeepEqual(got1, tt.wantEdges) {
				t.Errorf("createNetwork() got1 = %v, wantEdges %v", got1, tt.wantEdges)
			}
		})
	}
}
