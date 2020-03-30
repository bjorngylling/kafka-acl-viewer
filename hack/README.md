# Developer environment

## Build the application
The application is built in Docker and produces a Docker image as a result of the build. To build it locally simply
run `docker build -t kafka-acl-viewer:latest .` in the top directory.

If you're going to run it in Kubernetes as described below you will want to point to the Docker Daemon inside minikube
before you build it so the resulting image is available to the minikube cluster. This can be achieved with the
`minikube docker-env` command (see its output for the exact command you need to run).

## Run the application in Kubernetes environment
The main way to develop kafka-acl-viewer is to run a Kafka cluster with ACLs to point the application at. The easiest
way to get a ACL protected Kafka cluster up and running is to use Strimzi on Kubernetes. 

In the _k8s-test-env_ folder you can find configuration and deployments to get going. Before you can apply these
manifests you need a Kubernetes cluster with Strimzi support. See the 
[Strimzi Quickstart](https://strimzi.io/quickstarts/minikube/) for more information how to set that up. You want to
skip the step *Provision the Apache Kafka cluster*, and instead provision your Kafka cluster using the manifest
provided in _k8s-test-env/kafka/kafka-persistent-single-tls.yaml_.

```
kubectl apply -f k8s-test-env/kafka/kafka-persistent-single-tls.yaml \
    && kubectl wait kafka/my-cluster --for=condition=Ready --timeout=300s -n kafka
```

Once you have a Kafka cluster running you can apply the example topics and ACLs provided as well as the
kafka-acl-viewer deployment and user ACLs.
```
kubectl apply -f k8s-test-env/kafka-acl-viewer
```