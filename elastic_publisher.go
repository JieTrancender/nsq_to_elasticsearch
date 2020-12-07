package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jehiah/go-strftime"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// ElasticPublisher elastic publisher structure
type ElasticPublisher struct {
	client        *elastic.Client
	idxName       string
	idxType       string
	ddAccessToken string
}

// NewElasticPublisher create elastic publisher
func NewElasticPublisher(indexName string, indexType string, addrs []string, username, password, ddAccessToken string) (*ElasticPublisher, error) {
	var err error
	publisher := &ElasticPublisher{
		idxName:       indexName,
		idxType:       indexType,
		ddAccessToken: ddAccessToken,
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

// DingDingReqText dingding req text structure
type DingDingReqText struct {
	Content string `json:"content"`
}

// DingDingReqBodyInfo dingding req body structure
type DingDingReqBodyInfo struct {
	MsgType string          `json:"msgtype"`
	Text    DingDingReqText `json:"text"`
}

func (factory *ElasticPublisher) sendDingDingMsg(msg string) {
	client := &http.Client{}
	url := "oapi.dingtalk.com/robot/send"
	reqBody := DingDingReqBodyInfo{
		MsgType: "text",
		Text: DingDingReqText{
			Content: msg,
		},
	}

	reqBodyJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s?access_token=%s", url, factory.ddAccessToken), bytes.NewReader(reqBodyJSON))
	if err != nil {
		log.Printf("sendDingDingMsg fail:%v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("sendDingDingMsg success:%s\n", string(body))
	}
}

// 过滤报错信息并发送到顶顶群
func (factory *ElasticPublisher) filterTraceback(msg string) {
	if !strings.Contains(msg, "stack traceback") {
		return
	}

	// 发送钉钉消息
	if factory.ddAccessToken != "" {
		go factory.sendDingDingMsg(msg)
	}
}

func (factory *ElasticPublisher) handleMessage(m *nsq.Message) error {
	// fmt.Println("handleMessage", factory.indexName(), factory.indexType(), string(m.Body))
	data := make(map[string]interface{})
	err := json.Unmarshal(m.Body, &data)

	factory.filterTraceback(data["message"].(string))

	if err != nil {
		fmt.Println("Unmarshal fail", err)
		return err
	}
	entry := factory.client.Index().Index(factory.indexName()).BodyJson(data)
	_, err = entry.Do(context.Background())
	return err
}
