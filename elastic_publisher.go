package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jehiah/go-strftime"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"
	"time"
)

// ElasticPublisher elastic publisher structure
type ElasticPublisher struct {
	client  *elastic.Client
	idxName string
	idxType string
}

// NewElasticPublisher create elastic publisher
func NewElasticPublisher(indexName string, indexType string, addrs []string, username, password string) (*ElasticPublisher, error) {
	var err error
	publisher := &ElasticPublisher{
		idxName: indexName,
		idxType: indexType,
	}

	optionFuncs := []elastic.ClientOptionFunc{elastic.SetURL(addrs...)}
	if username != "" {
		optionFuncs = append(optionFuncs, elastic.SetBasicAuth(username, password))
	}

	publisher.client, err = elastic.NewClient(optionFuncs...)
	return publisher, err
}

func (factory *ElasticPublisher) indexName() string {
	now := time.Now()
	return strftime.Format(factory.idxName, now)
}

func (factory *ElasticPublisher) indexType() string {
	return factory.idxType
}

// type Person struct {
// 	Name string `json:"name"`
// }

func (factory *ElasticPublisher) handleMessage(m *nsq.Message) error {
	// fmt.Println("handleMessage", factory.indexName(), factory.indexType(), string(m.Body))
	data := make(map[string]interface{})
	err := json.Unmarshal(m.Body, &data)
	if err != nil {
		fmt.Println("Unmarshal fail", err)
		return err
	}
	entry := factory.client.Index().Index(factory.indexName()).BodyJson(data)
	_, err = entry.Do(context.Background())
	return err
}
