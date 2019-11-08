package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"github.com/Shopify/sarama"
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
	brokers       string
	version       string
	verbose       bool
	tlsCAFile     string
	tlsCertFile   string
	tlsKeyFile    string
	listenAddr    string
	fetchInterval time.Duration
}

type graphTemplateData struct {
	Nodes template.JS
	Edges template.JS
}

func parseFlags() (opts cmdOpts) {
	flag.StringVar(&opts.brokers, "brokers", "", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&opts.version, "version", "2.2.0", "Kafka cluster version")
	flag.BoolVar(&opts.verbose, "verbose", false, "Sarama logging")
	flag.StringVar(&opts.tlsCAFile, "ca-file", "", "Certificate authority file")
	flag.StringVar(&opts.tlsCertFile, "cert-file", "", "Client certificate file")
	flag.StringVar(&opts.tlsKeyFile, "key-file", "", "Client key file")
	flag.StringVar(&opts.listenAddr, "listen-addr", ":8080", "Address to listen on for the web interface")
	flag.DurationVar(&opts.fetchInterval, "fetch-interval", 10*time.Minute, "The interval at which to update the ACLs from Kafka")
	flag.Parse()

	if len(opts.brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if opts.verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	return
}

/**
 * Construct a new admin client connected to the kafka cluster.
 */
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
		RootCAs: x509.NewCertPool(),
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

func fetchUserOps(client sarama.ClusterAdmin) map[string]visjs.UserOps {
	// Load all topic ACLs
	resourceAcls, err := client.ListAcls(sarama.AclFilter{
		ResourceType:   sarama.AclResourceTopic,
		PermissionType: sarama.AclPermissionAny,
		Operation:      sarama.AclOperationAny,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return parseResourceAcls(resourceAcls)
}

func parseResourceAcls(acls []sarama.ResourceAcls) map[string]visjs.UserOps {
	// Convert ACLs into a data structure that is easier to build a graph from
	users := map[string]visjs.UserOps{}
	for _, resAcl := range acls {
		for _, acl := range resAcl.Acls {
			userDn := strings.TrimPrefix(acl.Principal, "User:")
			if _, ok := users[userDn]; !ok {
				users[userDn] = visjs.UserOps{To: map[string]struct{}{}, From: map[string]struct{}{}}
			}
			u := users[userDn]
			if acl.PermissionType == sarama.AclPermissionAllow {
				switch acl.Operation {
				case sarama.AclOperationRead:
					u.From[resAcl.ResourceName] = struct{}{}
				case sarama.AclOperationWrite:
					u.To[resAcl.ResourceName] = struct{}{}
				case sarama.AclOperationAll:
				case sarama.AclOperationAny:
					u.From[resAcl.ResourceName] = struct{}{}
					u.To[resAcl.ResourceName] = struct{}{}
				}
			}
		}
	}
	return users
}

func loadData(client sarama.ClusterAdmin) ([]string, map[string]visjs.UserOps) {
	topics, err := client.ListTopics()
	if err != nil {
		log.Fatalln(err)
	}
	keys := make([]string, 0, len(topics))
	for k := range topics {
		keys = append(keys, k)
	}
	return keys, fetchUserOps(client)
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
			nodes, edges = visjs.CreateNetwork(loadData(client))
			log.Printf("fetched data from kafka, load_duration=%s", time.Now().Sub(start))
			time.Sleep(opts.fetchInterval)
		}
	}()

	tmpl, err := template.ParseFiles("page.html")
	if err != nil {
		log.Fatal(err)
	}

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
