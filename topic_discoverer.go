package main

import (
	"log"
	"os"
	"regexp"
	"time"
)

// TopicDiscovererFunc interface of topic discoverer
type TopicDiscovererFunc func(topic string) error

// TopicDiscovererConfig config of topic discoverer
type TopicDiscovererConfig struct {
	LookupdAddress []string
	NsqdAddress    []string
	Pattern        string
	Refresh        time.Duration
	Handler        TopicDiscovererFunc
	Logger         *log.Logger
}

// TopicDiscoverer struct of topic discoverer
type TopicDiscoverer struct {
	TopicDiscovererConfig
	topics   map[string]bool
	termChan chan os.Signal
	Logger   *log.Logger
}

// NewTopicDiscoverer make new topic discoverer
func newTopicDiscoverer(cfg TopicDiscovererConfig, logger *log.Logger) (*TopicDiscoverer, error) {
	r := &TopicDiscoverer{
		TopicDiscovererConfig: cfg,
		termChan:              make(chan os.Signal),
		topics:                make(map[string]bool),
		Logger:                logger,
	}

	if r.Logger == nil {
		r.Logger = log.New(os.Stdout, "[topic_discoverer]: ", log.LstdFlags)
	}

	return r, nil
}

func (discoverer *TopicDiscoverer) allowTopicName(pattern, name string) bool {
	match, err := regexp.MatchString(pattern, name)
	if err != nil {
		return false
	}

	return match
}

func (discoverer *TopicDiscoverer) syncTopics() {
	// newTopics, err := look
}

func (discoverer *TopicDiscoverer) run() {
	// var ticker <-chan time.Time
	// if len(discoverer.)
}
