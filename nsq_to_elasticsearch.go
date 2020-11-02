package main

import (
	"flag"
	"fmt"
	"github.com/mreiferson/go-options"
	"github.com/nsqio/go-nsq"
	// "strings"
	// "github.com/olivere/elastic"
	"log"
	"os"
	"os/signal"
	"syscall"
	// "sync"
	"time"
)

// ArrayFlags for multi flag value
type ArrayFlags []string

// Set implement for Flag.Set
func (arrayFlags *ArrayFlags) Set(value string) error {
	*arrayFlags = append(*arrayFlags, value)

	return nil
}

// String for Flag.String
func (arrayFlags *ArrayFlags) String() string {
	return fmt.Sprint(*arrayFlags)
}

// Get for Flag.Get
func (arrayFlags *ArrayFlags) Get() interface{} {
	return []string(*arrayFlags)
}

// VERSION verions of nsqToElasticsearch
const VERSION = "0.0.1"

func flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("nsqToElasticsearch", flag.ExitOnError)

	fs.Bool("version", false, "show version")
	fs.String("log-level", "info", "set log verbosity: debug, info, warn, error, or fatal")
	fs.String("log-prefix", "[nsqToElasticsearch]", "log message prefix")

	fs.String("channel", "nsqToElasticsearch", "nsq channel")
	fs.Int("max-in-flight", 200, "max number of messages to allow in flight")

	fs.String("output-dir", "/tmp", "directory to write output files to")
	fs.String("work-dir", "", "directory for in-progress files before moving to output-dir")
	fs.Duration("topic-refresh", time.Minute, "how frequently the topic list should be refreshed")
	fs.String("topic-pattern", "", "only log topics matching the following pattern")
	fs.String("index-name", "nsq-%Y.%m.%d", "elasticsearch index name (strftime format)")
	fs.String("index-type", "nsq", "elasticsearch index mapping")

	fs.Duration("sync-interval", 30*time.Second, "sync file to elasticsearch duration")
	fs.Int("publisher-num", 10, "number of concurrent publishers")

	fs.Duration("http-client-connect-timeout", 2*time.Second, "timeout for HTTP connect")
	fs.Duration("http-client-request-timeout", 5*time.Second, "timeout for HTTP request")

	elasticHTTPAddrs := ArrayFlags{}
	nsqdTCPAddrs := ArrayFlags{}
	lookupdHTTPAddrs := ArrayFlags{}
	topics := ArrayFlags{}
	consumerOpts := ArrayFlags{}
	fs.Var(&elasticHTTPAddrs, "elasticsearch-http-address", "elasticsearch http address (may be given multiple times)")
	fs.Var(&nsqdTCPAddrs, "nsqd-tcp-address", "nsqd TCP address (may be given multiple times)")
	fs.Var(&lookupdHTTPAddrs, "lookupd-http-address", "lookupd HTTP address (may be given multiple times)")
	fs.Var(&topics, "topic", "nsq topic (may be given multiple times)")
	fs.Var(&consumerOpts, "consumer-opt", "option to passthrough to nsq.Config (may be given multiple times, http://godoc.org/github.com/nsqio/go-nsq#Config)")

	return fs
}

func main() {
	fs := flagSet()
	fs.Parse(os.Args[1:])

	if args := fs.Args(); len(args) > 0 {
		log.Fatalf("unknown arguments: %s", args)
	}

	opts := NewOptions()
	options.Resolve(opts, fs, nil)

	// logger := log.New(os.Stdout, "[topic_discoverer]: ", log.LstdFlags)

	if fs.Lookup("version").Value.(flag.Getter).Get().(bool) {
		fmt.Printf("nsq_to_elasticsearch@v%s go-nsq@v%s\n", VERSION, nsq.VERSION)
	}

	if opts.Channel == "" {
		log.Fatal("--channel is required")
	}

	if opts.HTTPClientConnectTimeout <= 0 {
		log.Fatal("--http-client-connect-timeout should be positive")
	}

	if opts.HTTPClientRequestTimeout <= 0 {
		log.Fatal("--http-client-request-timeout should be positive")
	}

	if len(opts.NSQDTCPAddrs) == 0 && len(opts.NSQLookupdHTTPAddrs) == 0 {
		log.Fatal("--nsqd-tcp-address or --lookupd-http-address required")
	}

	if len(opts.NSQDTCPAddrs) != 0 && len(opts.NSQLookupdHTTPAddrs) != 0 {
		log.Fatal("use --nsqd-tcp-address or --lookupd-http-address not both")
	}

	if len(opts.Topics) == 0 && len(opts.TopicPattern) == 0 {
		log.Fatal("--topic or --topic-pattern required")
	}

	if len(opts.Topics) != 0 && len(opts.NSQLookupdHTTPAddrs) == 0 {
		log.Fatal("--lookupd-http-address must be specified when no --topic specified")
	}

	if opts.WorkDir == "" {
		opts.WorkDir = opts.OutputDir
	}

	cfg := nsq.NewConfig()
	cfgFlag := nsq.ConfigFlag{cfg}
	for _, opt := range opts.ConsumerOpts {
		cfgFlag.Set(opt)
	}
	cfg.UserAgent = fmt.Sprintf("nsq_to_elasticsearch/%s go-nsq/%s", VERSION, nsq.VERSION)
	cfg.MaxInFlight = opts.MaxInFlight

	hupChan := make(chan os.Signal)
	termChan := make(chan os.Signal)
	signal.Notify(hupChan, syscall.SIGHUP)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	discoverer := newTopicDiscoverer(opts, cfg, hupChan, termChan)
	discoverer.run()

	// if *showVersion {
	// 	fmt.Printf("nsqToElasticsearch v, go-nsq v%s\n", VERSION, nsq.VERSION)
	// }

	// // if *topic == "" || *channel == "" {
	// // 	log.Fatal("--topic or --channel required")
	// // }
	// if len(topics) == 0 {
	// 	log.Fatal("--topic required")
	// }

	// if len(lookupdHTTPAddrs) == 0 && len(nsqdTCPAddrs) == 0 {
	// 	log.Fatal("--lookupd-http-address or --nsqd-tcp-address required")
	// }

	// if len(lookupdHTTPAddrs) != 0 && len(nsqdTCPAddrs) != 0 {
	// 	log.Fatal("use --nsqd-tcp-address or --lookupd-http-address not both")
	// }

	// if len(elasticAddrs) == 0 {
	// 	log.Fatal("missing --elasticsearch-http")
	// }

	// hupChan := make(chan os.Signal)
	// termChan := make(chan os.Signal)
	// signal.Notify(hupChan, syscall.SIGHUP)
	// siganl.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	// discoverer := newTopicDiscoverer()
	// discoverer.run()

	// cfg := nsq.NewConfig()
	// flagSet := flag.NewFlagSet("", flag.ExitOnError)

	// flagSet.Var(&nsq.ConfigFlag{cfg}, "consumer-opt", "option to pass through to nsq.Consumer (may be given multiple times)")
	// flagSet.PrintDefaults()

	// println("HeartbeatInterval", cfg.HeartbeatInterval)
	// println("MaxAttempts", cfg.MaxAttempts)

	// // err := flagSet.Parse([]string{
	// // 	"--consumer-opt=heartbeat_interval,1s",
	// // 	"--consumer-opt=max_attempts,10",
	// // })
	// // if err != nil {
	// // 	panic(err.Error())
	// // }

	// println("HeartbeatInterval", cfg.HeartbeatInterval)
	// println("MaxAttempts", cfg.MaxAttempts)
}
