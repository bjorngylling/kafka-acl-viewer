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
		want graph.Graph
	}{
		{
			name: "read_write_topic",
			args: args{
				acls: []sarama.ResourceAcls{
					{
						Resource: topicResource("topic-consume"),
						Acls: []*sarama.Acl{
							acl("user-1", sarama.AclOperationRead),
						},
					},
					{
						Resource: topicResource("topic-produce"),
						Acls: []*sarama.Acl{
							acl("user-1", sarama.AclOperationWrite),
						},
					},
				},
			},
			want: graph.Graph{
				Nodes: map[string]*graph.Node{
					"topic-consume": {Name: "topic-consume", Edges: []*graph.Edge{{Target: "user-1", Operation: "Read"}}, Type: "topic"},
					"user-1":        {Name: "user-1", Edges: []*graph.Edge{{Target: "topic-produce", Operation: "Write"}}, Type: "user"},
					"topic-produce": {Name: "topic-produce", Edges: nil, Type: "topic"},
				},
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
		ResourceType:        sarama.AclResourceTopic,
		ResourceName:        name,
		ResourcePatternType: sarama.AclPatternLiteral,
	}
}
