package main

import (
	"github.com/Shopify/sarama"
	"github.com/bjorngylling/kafka-acl-viewer/visjs"
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
		want map[string]visjs.UserOps
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
			want: map[string]visjs.UserOps{"CN=abc": {
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
