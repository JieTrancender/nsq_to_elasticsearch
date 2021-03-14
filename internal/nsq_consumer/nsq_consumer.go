package nsq_consumer

import (
	"github.com/nsqio/go-nsq"
	"log"
	"os"

	"github.com/JieTrancender/nsq_to_elasticsearch/internal/nsq_publisher"
	"github.com/JieTrancender/nsq_to_elasticsearch/internal/nsq_options"
)

// NSQConsumer nsq consumer structure
type NSQConsumer struct {
	publisher *nsq_publisher.ElasticPublisher
	opts      *nsq_options.Options
	topic     string
	consumer  *nsq.Consumer

	msgChan chan *nsq.Message

	TermChan chan bool
	HupChan  chan bool
}

// NewNSQConsumer create NSQConsumer
func NewNSQConsumer(opts *nsq_options.Options, topic string, cfg *nsq.Config,
	elasticAddrs []string, idxName, idxType, elasticUsername, elasticPassword, ddAccessToken string) (*NSQConsumer, error) {
	log.Println("NewNSQConsumer topic", topic)
	publisher, err := nsq_publisher.NewElasticPublisher(idxName, idxType, elasticAddrs, elasticUsername, elasticPassword, ddAccessToken)
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
		TermChan:  make(chan bool),
		HupChan:   make(chan bool),
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

func (nsqConsumer *NSQConsumer) Router() {
	closeElastic, exit := false, false
	for {
		select {
		case <-nsqConsumer.consumer.StopChan:
			closeElastic, exit = true, true
		case <-nsqConsumer.TermChan:
			nsqConsumer.consumer.Stop()
		case <-nsqConsumer.HupChan:
			closeElastic = true
		case m := <-nsqConsumer.msgChan:
			err := nsqConsumer.publisher.HandleMessage(m)
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
