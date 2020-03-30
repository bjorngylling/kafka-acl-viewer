package visjs

import (
	"github.com/bjorngylling/kafka-acl-viewer/graph"
	"reflect"
	"sort"
	"strings"
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
						"user-1":  {"user-1", "user", nil},
						"topic-1": {"topic-1", "topic", []*graph.Edge{{Target: "user-1", Operation: "Read"}}},
					},
				},
			},
			wantNodes: []Node{
				{
					ID:    "user-1",
					Label: "🤖 user-1",
					Shape: "box",
					Color: color{"#6ef091", highlight{Background: "#ccffda"}},
					Type:  "user",
				},
				{
					ID:    "topic-1",
					Label: "🗒 topic-1",
					Shape: "box",
					Color: color{"", highlight{""}},
					Type:  "topic",
				},
			},
			wantEdges: []Edge{{From: "topic-1", To: "user-1", Arrows: "to", Dashes: false, Title: "Read"}},
		},
		{
			name: "cluster",
			args: args{
				graph: graph.Graph{
					Nodes: map[string]*graph.Node{
						"user-1":        {"user-1", "user", nil},
						"Kafka Cluster": {"Kafka Cluster", "cluster", []*graph.Edge{{Target: "user-1", Operation: "Describe"}}},
					},
				},
			},
			wantNodes: []Node{
				{
					ID:    "user-1",
					Label: "🤖 user-1",
					Shape: "box",
					Color: color{"#6ef091", highlight{Background: "#ccffda"}},
					Type:  "user",
				},
				{
					ID:    "Kafka Cluster",
					Label: "Kafka Cluster",
					Shape: "box",
					Color: color{"", highlight{""}},
					Type:  "cluster",
				},
			},
			wantEdges: []Edge{{From: "Kafka Cluster", To: "user-1", Arrows: "to", Dashes: false, Title: "Describe"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CreateNetwork(tt.args.graph)

			// sort to since the order is not important
			sort.Slice(got, func(i, j int) bool {
				return strings.Compare(got[i].ID, got[j].ID) < 0
			})
			sort.Slice(tt.wantNodes, func(i, j int) bool {
				return strings.Compare(tt.wantNodes[i].ID, tt.wantNodes[j].ID) < 0
			})

			if !reflect.DeepEqual(got, tt.wantNodes) {
				t.Errorf("createNetwork() got = %v, wantNodes %v", got, tt.wantNodes)
			}
			if !reflect.DeepEqual(got1, tt.wantEdges) {
				t.Errorf("createNetwork() got1 = %v, wantEdges %v", got1, tt.wantEdges)
			}
		})
	}
}
