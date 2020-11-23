package main

import (
	"github.com/nsqio/go-nsq"
	"log"
	"os"
)

// NSQConsumer nsq consumer structure
type NSQConsumer struct {
	publisher *ElasticPublisher
	opts      *Options
	topic     string
	consumer  *nsq.Consumer

	msgChan chan *nsq.Message

	termChan chan bool
	hupChan  chan bool
}

// NewNSQConsumer create NSQConsumer
func NewNSQConsumer(opts *Options, topic string, cfg *nsq.Config,
	elasticAddrs []string, idxName, idxType, elasticUsername, elasticPassword string) (*NSQConsumer, error) {
	log.Println("NewNSQConsumer topic", topic)
	publisher, err := NewElasticPublisher(idxName, idxType, elasticAddrs, elasticUsername, elasticPassword)
	if err != nil {
		return nil, err
	}

	consumer, err := nsq.NewConsumer(topic, opts.Channel, cfg)
	if err != nil {
		return nil, err
	}

	nsqConsumer := &NSQConsumer{
		publisher: publisher,
		opts:      opts,
		topic:     topic,
		consumer:  consumer,
		msgChan:   make(chan *nsq.Message, 1),
		termChan:  make(chan bool),
		hupChan:   make(chan bool),
	}
	consumer.AddHandler(nsqConsumer)

	err = consumer.ConnectToNSQDs(opts.NSQDTCPAddrs)
	if err != nil {
		return nil, err
	}

	err = consumer.ConnectToNSQLookupds(opts.NSQLookupdHTTPAddrs)
	if err != nil {
		return nil, err
	}

	return nsqConsumer, nil
}

// HandleMessage implement of NSQ HandleMessage interface
func (nsqConsumer *NSQConsumer) HandleMessage(m *nsq.Message) error {
	m.DisableAutoResponse()
	nsqConsumer.msgChan <- m
	return nil
}

func (nsqConsumer *NSQConsumer) router() {
	closeElastic, exit := false, false
	for {
		select {
		case <-nsqConsumer.consumer.StopChan:
			closeElastic, exit = true, true
		case <-nsqConsumer.termChan:
			nsqConsumer.consumer.Stop()
		case <-nsqConsumer.hupChan:
			closeElastic = true
		case m := <-nsqConsumer.msgChan:
			err := nsqConsumer.publisher.handleMessage(m)
			if err != nil {
				// 重试
				m.Requeue(-1)
				log.Println("NSQConsumer router msg deal fail", err)
				os.Exit(1)
			}
			m.Finish()
		}

		if closeElastic {
			nsqConsumer.Close()
			closeElastic = false
		}

		if exit {
			break
		}
	}
}

// Close close this NSQConsumer
func (nsqConsumer *NSQConsumer) Close() {
	log.Println("NSQConsumer Close")
}
