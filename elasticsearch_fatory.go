package main

// // ElasticFactory for elasticsearch
// type ElasticFactory struct {
// 	idxName        string
// 	idxType        string
// 	metricsTimeout int
// 	elasticAddrs   []string
// 	wg             sync.WaitGroup
// 	mtx            sync.Mutex
// 	consumers      []*nsq.Consumer
// }

// func newElasticFactory() (*ElasticFactory, error) {
// 	return &ElasticFactory{}, nil
// }

// // Signal for system signal
// func (f *ElasticFactory) Signal(sig os.Signal) {
// 	f.Stop()
// }

// func (f *ElasticFactory) startConsumer(consumer *nsq.Consumer) {
// 	f.wg.Add(1)
// 	defer f.wg.Done()
// 	<-consumer.StopChan
// }

// // RegisterTopic register nsq topic
// func (f *ElasticFactory) RegisterTopic(name string) error {
// 	f.mtx.Lock()
// 	defer f.mtx.Unlock()

// 	log.Println("Registering topic", name)
// 	publisher, err := newElasticPublisher(*indexName, *indexType, *statusEvery, []string(elasticAddrs))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	cfg := nsq.NewConfig()
// 	cfgFlag := nsq.ConfigFlag{cfg}
// 	for _, opt := range consumerOpts {
// 		cfgFlag.Set(opt)
// 	}
// 	cfg.UserAgent = fmt.Sprintf("nsqToElasticsearch/%s go-nsq/%s", VERSION, nsq.VERSION)
// 	cfg.MaxInFlight = *maxInFlight

// 	consumer, err := nsq.NewConsumer(name, *channel, cfg)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	consumer.AddConcurrentHandlers(publisher, *numPublishers)
// 	err = consumer.ConnecToNSQDs(nsqdTCPAddrs)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = consumer.ConnectToNSQLookupds(lookupdHTTPAddrs)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	f.consumers = append(f.consumers, consumer)
// 	go f.startConsumer(consumer)

// 	return nil
// }
