package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"github.com/Shopify/sarama"
	"github.com/bjorngylling/kafka-acl-viewer/graph"
	"github.com/bjorngylling/kafka-acl-viewer/visjs"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type cmdOpts struct {
	brokers            string
	version            string
	verbose            bool
	tlsCAFile          string
	tlsCertFile        string
	tlsKeyFile         string
	listenAddr         string
	fetchInterval      time.Duration
	insecureSkipVerify bool
}

type graphTemplateData struct {
	Nodes template.JS
	Edges template.JS
}

func parseFlags() cmdOpts {
	opts := cmdOpts{}
	flag.StringVar(&opts.brokers, "brokers", lookupEnvOrString("KAFKA_URL", ""), "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&opts.version, "version", lookupEnvOrString("KAFKA_VERSION", "2.2.0"), "Kafka cluster version")
	flag.BoolVar(&opts.verbose, "verbose", false, "Sarama logging")
	flag.StringVar(&opts.tlsCAFile, "ca-file", lookupEnvOrString("CA_FILE", ""), "Certificate authority file")
	flag.StringVar(&opts.tlsCertFile, "cert-file", lookupEnvOrString("CERT_FILE", ""), "Client certificate file")
	flag.StringVar(&opts.tlsKeyFile, "key-file", lookupEnvOrString("KEY_FILE", ""), "Client key file")
	flag.StringVar(&opts.listenAddr, "listen-addr", lookupEnvOrString("LISTEN_ADDR", ":8080"), "Address to listen on for the web interface")
	flag.BoolVar(&opts.insecureSkipVerify, "insecure-skip-verify", false, "Skip hostname verification in the TLS handshake, ONLY USE THIS IF YOU KNOW WHAT IT MEANS")
	flag.DurationVar(&opts.fetchInterval, "fetch-interval", lookupEnvOrDuration("FETCH_INTERVAL", 10*time.Minute), "The interval at which to update the ACLs from Kafka")
	flag.Parse()

	if len(opts.brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if opts.verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	return opts
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		dur, err := time.ParseDuration(val)
		if err != nil {
			log.Fatalf("parse duration from env[%s] failed: %v", key, err)
		}
		return dur
	}
	return defaultVal
}

// Construct a new admin client connected to the kafka cluster.
func createAdminClient(opts cmdOpts) sarama.ClusterAdmin {
	config := sarama.NewConfig()

	version, err := sarama.ParseKafkaVersion(opts.version)
	if err != nil {
		log.Fatalln(err)
	}
	config.Version = version
	config.ClientID = "kafka-acl-viewer"

	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: opts.insecureSkipVerify,
		RootCAs:            x509.NewCertPool(),
	}
	if ca, err := ioutil.ReadFile(opts.tlsCAFile); err == nil {
		config.Net.TLS.Config.RootCAs.AppendCertsFromPEM(ca)
	} else {
		log.Fatalln(err)
	}

	cert, err := tls.LoadX509KeyPair(opts.tlsCertFile, opts.tlsKeyFile)
	if err == nil {
		config.Net.TLS.Config.Certificates = []tls.Certificate{cert}
	} else {
		log.Fatalln(err)
	}

	brokers := strings.Split(opts.brokers, ",")
	client, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

func loadAclGraph(client sarama.ClusterAdmin) graph.Graph {
	resourceAcls, err := client.ListAcls(sarama.AclFilter{
		ResourceType:              sarama.AclResourceAny,
		PermissionType:            sarama.AclPermissionAny,
		Operation:                 sarama.AclOperationAny,
		ResourcePatternTypeFilter: sarama.AclPatternAny,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return parseResourceAcls(resourceAcls)
}

func parseResourceAcls(acls []sarama.ResourceAcls) graph.Graph {
	aclGraph := graph.NewGraph()
	for _, resAcl := range acls {
		resourceName := resAcl.ResourceName
		resourceType := "topic"
		if resAcl.ResourceType == sarama.AclResourceGroup {
			continue // we dont display consumer groups at this point
		}
		if resAcl.ResourceType == sarama.AclResourceCluster {
			resourceName = "Kafka Cluster"
			resourceType = "cluster"
		}
		if _, ok := aclGraph.Nodes[resourceName]; !ok {
			aclGraph.AddNode(&graph.Node{
				Name:  resourceName,
				Type:  resourceType,
				Edges: nil,
			})
		}
		resource := aclGraph.Nodes[resourceName]
		for _, acl := range resAcl.Acls {
			userDn := strings.TrimPrefix(acl.Principal, "User:")
			if _, ok := aclGraph.Nodes[userDn]; !ok {
				aclGraph.AddNode(&graph.Node{
					Name:  userDn,
					Type:  "user",
					Edges: nil,
				})
			}
			user := aclGraph.Nodes[userDn]
			if acl.PermissionType == sarama.AclPermissionAllow {
				if resAcl.ResourceType == sarama.AclResourceCluster {
					switch acl.Operation {
					case sarama.AclOperationDescribe:
						resource.AddEdge(&graph.Edge{Target: user.Name, Operation: stringify(acl.Operation)})
					case sarama.AclOperationAlter:
						user.AddEdge(&graph.Edge{Target: resourceName, Operation: stringify(acl.Operation)})
					}
				} else if resAcl.ResourceType == sarama.AclResourceTopic {
					switch acl.Operation {
					case sarama.AclOperationRead:
						resource.AddEdge(&graph.Edge{Target: user.Name, Operation: stringify(acl.Operation)})
					case sarama.AclOperationWrite:
						user.AddEdge(&graph.Edge{Target: resourceName, Operation: stringify(acl.Operation)})
					case sarama.AclOperationAll:
					case sarama.AclOperationAny:
						resource.AddEdge(&graph.Edge{Target: user.Name, Operation: "Read"})
						user.AddEdge(&graph.Edge{Target: resourceName, Operation: "Write"})
					}
				}
			}
		}
	}
	return aclGraph
}

func stringify(op sarama.AclOperation) string {
	switch op {
	case sarama.AclOperationAlter:
		return "Alter"
	case sarama.AclOperationDescribe:
		return "Describe"
	case sarama.AclOperationRead:
		return "Read"
	case sarama.AclOperationWrite:
		return "Write"
	}
	return ""
}

func main() {
	opts := parseFlags()
	log.Printf("connecting to kafka brokers=%s", opts.brokers)
	client := createAdminClient(opts)

	var nodes []visjs.Node
	var edges []visjs.Edge
	go func() {
		for {
			start := time.Now()
			nodes, edges = visjs.CreateNetwork(loadAclGraph(client))
			log.Printf("fetched data from kafka, load_duration=%s", time.Now().Sub(start))
			time.Sleep(opts.fetchInterval)
		}
	}()

	tmpl, err := template.ParseFiles("web/page.html")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix(strings.TrimRight("/static/", "/"), fileServer))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		jsonNodes, err := json.Marshal(nodes)
		if err != nil {
			log.Fatalln(err)
		}
		jsonEdges, err := json.Marshal(edges)
		if err != nil {
			log.Fatalln(err)
		}
		err = tmpl.Execute(w, graphTemplateData{template.JS(string(jsonNodes)), template.JS(string(jsonEdges))})
		if err != nil {
			log.Fatalln(err)
		}
	})

	log.Println("listening on", opts.listenAddr)
	log.Fatal(http.ListenAndServe(opts.listenAddr, nil))
}
