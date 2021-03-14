package topic_discoverer

import (
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"

	"github.com/JieTrancender/nsq_to_elasticsearch/internal/nsq_options"
	"github.com/JieTrancender/nsq_to_elasticsearch/internal/nsq_consumer"
)

// TopicDiscoverer struct of topic discoverer
type TopicDiscoverer struct {
	opts            *nsq_options.Options
	topics          map[string]*(nsq_consumer.NSQConsumer)
	termChan        chan os.Signal
	hupChan         chan os.Signal
	logger          *log.Logger
	wg              sync.WaitGroup
	cfg             *nsq.Config
	elasticAddrs    []string
	idxName         string
	idxType         string
	elasticUserName string
	elasticPassword string
	ddAccessToken   string
}

func NewTopicDiscoverer(opts *nsq_options.Options, cfg *nsq.Config, hupChan chan os.Signal, termChan chan os.Signal,
	ddAccessToken string) (*TopicDiscoverer, error) {
	discoverer := &TopicDiscoverer{
		opts:            opts,
		topics:          make(map[string]*nsq_consumer.NSQConsumer),
		termChan:        termChan,
		hupChan:         hupChan,
		logger:          log.New(os.Stdout, "[topic_discoverer]: ", log.LstdFlags),
		cfg:             cfg,
		ddAccessToken:   ddAccessToken,
	}

	return discoverer, nil
}

// func (discoverer *TopicDiscoverer) allowTopicName(pattern, name string) bool {
// 	match, err := regexp.MatchString(pattern, name)
// 	if err != nil {
// 		return false
// 	}

// 	return match
// }

func (discoverer *TopicDiscoverer) isTopicAllowed(topic string) bool {
	if len(discoverer.opts.TopicPatterns) == 0 {
		return true
	}

	var match bool
	var err error
	for _, pattern := range discoverer.opts.TopicPatterns {
		match, err = regexp.MatchString(pattern, topic)
		if err == nil {
			break
		}
	}

	return match
}

func (discoverer *TopicDiscoverer) updateTopics(topics []string) {
	for _, topic := range topics {
		if _, ok := discoverer.topics[topic]; ok {
			continue
		}

		if !discoverer.isTopicAllowed(topic) {
			discoverer.logger.Printf("skipping topic %s (doesn't match any pattern)\n", topic)
			continue
		}

		nsqConsumer, err := nsq_consumer.NewNSQConsumer(discoverer.opts, topic, discoverer.cfg,
			discoverer.elasticAddrs, discoverer.idxName, discoverer.idxType,
			discoverer.elasticUserName, discoverer.elasticPassword, discoverer.ddAccessToken)
		if err != nil {
			discoverer.logger.Printf("error: could not register topic %s: %s", topic, err)
			continue
		}
		discoverer.topics[topic] = nsqConsumer

		discoverer.wg.Add(1)
		go func(nsqConsumer *nsq_consumer.NSQConsumer) {
			nsqConsumer.Router()
			discoverer.wg.Done()
		}(nsqConsumer)
	}
}

func (discoverer *TopicDiscoverer) Run() {
	var ticker <-chan time.Time
	if len(discoverer.opts.Topics) == 0 {
		ticker = time.Tick(discoverer.opts.TopicRefreshInterval)
	}
	discoverer.updateTopics(discoverer.opts.Topics)

forloop:
	for {
		select {
		case <-ticker:
			discoverer.updateTopics(discoverer.opts.Topics)
		case <-discoverer.termChan:
			for _, nsqConsumer := range discoverer.topics {
				// nsqConsumer.consumer.Stop()
				close(nsqConsumer.TermChan)
			}
			break forloop
		case <-discoverer.hupChan:
			for _, nsqConsumer := range discoverer.topics {
				nsqConsumer.HupChan <- true
			}
			break forloop
		}
	}

	discoverer.wg.Wait()
}
