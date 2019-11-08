package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"github.com/Shopify/sarama"
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

type nodeShape string

var (
	ShapeCircle nodeShape = "circle"
	ShapeBox    nodeShape = "box"

	ColorGreen = color{Background: "#6ef091", Highlight: highlight{Background: "#ccffda"}}
)

type highlight struct {
	Background string `json:"background"`
}
type color struct {
	Background string    `json:"background"`
	Highlight  highlight `json:"highlight"`
}

type node struct {
	ID    int       `json:"id"`
	Label string    `json:"label"`
	Shape nodeShape `json:"shape"`
	Color color     `json:"color"`
}

type edge struct {
	From   int    `json:"from"`
	To     int    `json:"to"`
	Arrows string `json:"arrows"`
	Dashes bool   `json:"dashes,omitempty"`
	Title  string `json:"title,omitempty"`
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

type userOps struct {
	To   map[string]struct{}
	From map[string]struct{}
}

func fetchUserOps(client sarama.ClusterAdmin) map[string]userOps {
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

func parseResourceAcls(acls []sarama.ResourceAcls) map[string]userOps {
	// Convert ACLs into a data structure that is easier to build a graph from
	users := map[string]userOps{}
	for _, resAcl := range acls {
		for _, acl := range resAcl.Acls {
			userDn := strings.TrimPrefix(acl.Principal, "User:")
			if _, ok := users[userDn]; !ok {
				users[userDn] = userOps{To: map[string]struct{}{}, From: map[string]struct{}{}}
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

func loadData(client sarama.ClusterAdmin) ([]node, []edge) {
	users := fetchUserOps(client)

	// Create user nodes, i.e. consumers and producers
	var nodes []node
	i := 0
	userIdLookup := map[string]int{}
	for user := range users {
		userIdLookup[user] = i
		nodes = append(nodes, node{
			ID:    i,
			Label: "🤖 " + user,
			Shape: ShapeBox,
			Color: ColorGreen,
		})
		i++
	}

	// Create topic nodes
	topics, err := client.ListTopics()
	if err != nil {
		log.Fatalln(err)
	}
	topicIdLookup := map[string]int{}
	for topic := range topics {
		topicIdLookup[topic] = i
		nodes = append(nodes, node{
			ID:    i,
			Label: "🗒 " + topic,
			Shape: ShapeBox,
		})
		i++
	}

	// Add all the edges
	var edges []edge
	for user, ops := range users {
		for input := range ops.From {
			edges = append(edges, edge{
				From:   topicIdLookup[input],
				To:     userIdLookup[user],
				Arrows: "to",
				Dashes: false,
				Title:  "Read",
			})
		}
		for output := range ops.To {
			edges = append(edges, edge{
				From:   userIdLookup[user],
				To:     topicIdLookup[output],
				Arrows: "to",
				Dashes: false,
				Title:  "Write",
			})
		}
	}

	return nodes, edges
}

func main() {
	opts := parseFlags()
	log.Printf("connecting to kafka brokers=%s", opts.brokers)
	client := createAdminClient(opts)

	var nodes []node
	var edges []edge
	go func() {
		for {
			start := time.Now()
			nodes, edges = loadData(client)
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
