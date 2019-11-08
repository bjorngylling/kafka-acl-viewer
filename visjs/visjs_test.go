package visjs

import (
	"reflect"
	"testing"
)

func Test_createNetwork(t *testing.T) {
	type args struct {
		topics         []string
		userOperations map[string]UserOps
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
				topics: []string{"topic-1"},
				userOperations: map[string]UserOps{
					"CN=abc": {
						To: map[string]struct{}{},
						From: map[string]struct{}{
							"topic-1": {},
						},
					},
				},
			},
			wantNodes: []Node{
				{
					ID:    0,
					Label: "🤖 CN=abc",
					Shape: "box",
					Color: color{"#6ef091", highlight{Background: "#ccffda"}},
				},
				{
					ID:    1,
					Label: "🗒 topic-1",
					Shape: "box",
					Color: color{"", highlight{""}},
				},
			},
			wantEdges: []Edge{
				{
					From:   1,
					To:     0,
					Arrows: "to",
					Dashes: false,
					Title:  "Read",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CreateNetwork(tt.args.topics, tt.args.userOperations)
			if !reflect.DeepEqual(got, tt.wantNodes) {
				t.Errorf("createNetwork() got = %v, wantNodes %v", got, tt.wantNodes)
			}
			if !reflect.DeepEqual(got1, tt.wantEdges) {
				t.Errorf("createNetwork() got1 = %v, wantEdges %v", got1, tt.wantEdges)
			}
		})
	}
}
