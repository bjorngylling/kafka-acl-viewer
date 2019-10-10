# Kafka ACL Viewer
An application to display your Kafka cluster ACLs as a graph of topics and consumers/producers. It allows you to
visualise how events flow within your cluster in a simple way.

## Authentication with Kafka
This application is built to authenticate against Kafka using TLS. It requires two ACL entries on the broker to
function properly. The first one, to read ACLs from the cluster itself `--cluster --operation read` and to describe
all topics in the cluster, `--topic='*' --operation describe`.
