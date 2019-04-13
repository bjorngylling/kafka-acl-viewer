package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type kafkaOpts struct {
	brokers     string
	version     string
	verbose     bool
	tlsCAFile   string
	tlsCertFile string
	tlsKeyFile  string
}

func main() {
	opts := kafkaOpts{}
	flag.StringVar(&opts.brokers, "brokers", "", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&opts.version, "version", "2.2.0", "Kafka cluster version")
	flag.BoolVar(&opts.verbose, "verbose", false, "Sarama logging")
	flag.StringVar(&opts.tlsCAFile, "ca-file", "", "Certificate authority file")
	flag.StringVar(&opts.tlsCertFile, "cert-file", "", "Client certificate file")
	flag.StringVar(&opts.tlsKeyFile, "key-file", "", "Client key file")
	flag.Parse()

	if len(opts.brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if opts.verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(opts.version)
	if err != nil {
		log.Fatalln(err)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
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

	resourceAcls, err := client.ListAcls(sarama.AclFilter{
		ResourceType:   sarama.AclResourceTopic,
		PermissionType: sarama.AclPermissionAny,
		Operation:      sarama.AclOperationAny,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%#v\n", resourceAcls)
}
