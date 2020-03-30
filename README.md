# Kafka ACL Viewer
[![Go Report Card](https://goreportcard.com/badge/github.com/bjorngylling/kafka-acl-viewer)](https://goreportcard.com/report/github.com/bjorngylling/kafka-acl-viewer) [![Build](https://github.com/bjorngylling/kafka-acl-viewer/workflows/build/badge.svg)](https://github.com/bjorngylling/kafka-acl-viewer/actions)

An application to display your Kafka cluster ACLs as a graph of topics and consumers/producers. It allows you to
visualise how events flow within your cluster in a simple way. See the picture below for an example of how it can look.

<img src="kafka-acl-viewer.png?raw=true" width="600" title="Kafka ACL viewer in action">

**Note:** currently all operation types are not displayed and resources (read topics) may not always be "real" topics, the ACL may be a
wildcard resource such as `*` for all topics. Don't use this data as the absolute truth of what accesses are in your cluster.

## Authentication with Kafka
This application is built to authenticate against Kafka using TLS. It requires two ACL entries on the broker to
function properly. The first one, to read ACLs from the cluster itself `--cluster --operation read` and to describe
all topics in the cluster, `--topic='*' --operation describe`.
