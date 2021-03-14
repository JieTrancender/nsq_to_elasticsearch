package nsq_publisher

import (
	"flag"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"
)

var (
	elasticHTTPAddrs = make([]string, 1)
	// elasticHTTPAddrs = flag.StringArray("elastic_http_addrs", []string{}, "http addrs for elasticsearch")
	elasticUsername = flag.String("elastic_username", "", "username for elasticsearch")
	elasticPassword = flag.String("elastic_password", "", "password for elasticsearch")
	elasticIndexName = flag.String("elastic_index_name", "nsq-%Y.%m.%d", "elasticsearch index name (strftime format")
	elasticIndexType = flag.String("elastic_index_type", "nsq", "elasticsearch index mapping")
)

// ElasticPublisher elastic publisher structure
type ElasticPublisher struct {
	client        *elastic.Client
	idxName       string
	idxType       string
	ddAccessToken string
}

// NewElasticPublisher create elastic publisher
func NewElasticPublisher(ddAccessToken string) (*ElasticPublisher, error) {
	var err error
	publisher := &ElasticPublisher{
		idxType: *elasticIndexType,
		idxName: *elasticIndexName,
		ddAccessToken: ddAccessToken,
	}

	optionFuncs := []elastic.ClientOptionFunc{elastic.SetURL(elasticHTTPAddrs...)}
	if *elasticUsername != "" {
		optionFuncs = append(optionFuncs, elastic.SetBasicAuth(*elasticUsername, *elasticPassword))
	}

	publisher.client, err = elastic.NewClient(optionFuncs...)
	return publisher, err
}

func (factory *ElasticPublisher) indexName() string {
	now := time.Now()
	return strftime.Format(factory.idxName, now)
}

// func (factory *ElasticPublisher) indexType() string {
// 	return factory.idxType
// }

// type Person struct {
// 	Name string `json:"name"`
// }

// DingDingReqText dingding req text structure
// type DingDingReqText struct {
// 	Content string `json:"content"`
// }

// DingDingReqMarkdown request for Markdown format
type DingDingReqMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingDingReqBodyInfo dingding req body structure
type DingDingReqBodyInfo struct {
	MsgType string `json:"msgtype"`
	// Text    DingDingReqText `json:"text"`
	Markdown DingDingReqMarkdown `json:"markdown"`
}

// LogDataInfo log data structure
type LogDataInfo struct {
	GamePlatform string `json:"gamePlatform"`
	NodeName     string `json:"nodeName"`
	FileName     string `json:"fileName"`
	Msg          string `json:"message"`
}

func generateMarkDownBody(logData LogDataInfo) ([]byte, error) {
	reqBody := DingDingReqBodyInfo{
		MsgType: "markdown",
		Markdown: DingDingReqMarkdown{
			Title: "报错信息",
			Text:  fmt.Sprintf("\n\n## %s渠道%s节点报错收集\n\n文件名:**%s**\n\n```lua\n%s\n```", logData.GamePlatform, logData.NodeName, logData.FileName, logData.Msg),
		},
	}

	return json.Marshal(reqBody)
}

func (factory *ElasticPublisher) sendDingDingMsg(logData LogDataInfo) {
	client := &http.Client{}
	url := "oapi.dingtalk.com/robot/send"
	reqBodyJSON, err := generateMarkDownBody(logData)
	if err != nil {
		log.Printf("generate req data fail:%s", err)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s?access_token=%s", url, factory.ddAccessToken), bytes.NewReader(reqBodyJSON))
	if err != nil {
		fmt.Printf("sendDingDingMsg fail:%v %s", err, string(reqBodyJSON))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("sendDingDingMsg do fail:%v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("sendDingDingMsg success:%s\n", string(body))
	}
}

// 过滤报错信息并发送到顶顶群
func (factory *ElasticPublisher) filterTraceback(gamePlatform, nodeName, fileName, msg string) {
	// 报错堆栈或者alarm日志才报警
	// todo: 使用可配置的关键字
	if !strings.Contains(msg, "stack traceback") && !strings.Contains(msg, "alarm") {
		return
	}

	// 特定关键字不报警，例如聊天后台请求
	// todo: 使用可配置的关键字
	if strings.Contains(msg, "chatMsgFilter") || strings.Contains(msg, "attempt to index a nil value (local 'pb')") ||
		strings.Contains(msg, "websocket") {
		return
	}

	if true {
		return
	}

	// 发送钉钉消息
	if factory.ddAccessToken != "" {
		logData := LogDataInfo{
			GamePlatform: gamePlatform,
			NodeName:     nodeName,
			FileName:     fileName,
			Msg:          msg,
		}
		go factory.sendDingDingMsg(logData)
	}
}

func (factory *ElasticPublisher) HandleMessage(m *nsq.Message) error {
	// fmt.Println("handleMessage", factory.indexName(), factory.indexType(), string(m.Body))
	data := make(map[string]interface{})
	err := json.Unmarshal(m.Body, &data)
	if err != nil {
		fmt.Println("Unmarshal fail", err)
		return err
	}

	logData := data["log"].(map[string]interface{})
	fileData := logData["file"].(map[string]interface{})
	factory.filterTraceback(data["gamePlatform"].(string), data["nodeName"].(string), fileData["path"].(string), data["message"].(string))

	entry := factory.client.Index().Index(factory.indexName()).BodyJson(data)
	_, err = entry.Do(context.Background())
	return err
}
