package main

import (
	"github.com/Shopify/sarama"
	"github.com/bjorngylling/kafka-acl-viewer/graph"
	"reflect"
	"testing"
)

func Test_parseResourceAcls(t *testing.T) {
	type args struct {
		acls []sarama.ResourceAcls
	}
	tests := []struct {
		name string
		args args
		want map[string]userOps
	}{
		{
			name: "read_write_topic",
			args: args{
				acls: []sarama.ResourceAcls{
					{
						Resource: topicResource("topic-consume"),
						Acls: []*sarama.Acl{
							acl("CN=abc", sarama.AclOperationRead),
						},
					},
					{
						Resource: topicResource("topic-produce"),
						Acls: []*sarama.Acl{
							acl("CN=abc", sarama.AclOperationWrite),
						},
					},
				},
			},
			want: map[string]userOps{"CN=abc": {
				To: map[string]struct{}{
					"topic-produce": {},
				},
				From: map[string]struct{}{
					"topic-consume": {},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseResourceAcls(tt.args.acls); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResourceAcls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func acl(principal string, operation sarama.AclOperation) *sarama.Acl {
	return &sarama.Acl{
		Principal:      principal,
		Host:           "*",
		Operation:      operation,
		PermissionType: sarama.AclPermissionAllow,
	}
}

func topicResource(name string) sarama.Resource {
	return sarama.Resource{
		ResourceType:       sarama.AclResourceTopic,
		ResourceName:       name,
		ResoucePatternType: sarama.AclPatternLiteral,
	}
}

func Test_createGraph(t *testing.T) {
	type args struct {
		topics []string
		ops    map[string]userOps
	}
	tests := []struct {
		name string
		args args
		want graph.Graph
	}{
		{
			name: "read_topic",
			args: args{
				topics: []string{"topic-1"},
				ops: map[string]userOps{"user-1": {
					From: map[string]struct{}{
						"topic-1": {},
					}},
				},
			},
			want: graph.Graph{
				Nodes: map[string]*graph.Node{
					"user-1":  {Name: "user-1", Edges: []*graph.Edge{}, Type: "user"},
					"topic-1": {Name: "topic-1", Edges: []*graph.Edge{{Target: "user-1", Operation: "Read"}}, Type: "topic"},
				},
			},
		},
		{
			name: "write_topic",
			args: args{
				topics: []string{"topic-1"},
				ops: map[string]userOps{"user-1": {
					To: map[string]struct{}{
						"topic-1": {},
					}},
				},
			},
			want: graph.Graph{
				Nodes: map[string]*graph.Node{
					"user-1":  {Name: "user-1", Edges: []*graph.Edge{{Target: "topic-1", Operation: "Write"}}, Type: "user"},
					"topic-1": {Name: "topic-1", Edges: []*graph.Edge{}, Type: "topic"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createGraph(tt.args.topics, tt.args.ops); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createGraph() = %v, want %v", got, tt.want)
			}
		})
	}
}
