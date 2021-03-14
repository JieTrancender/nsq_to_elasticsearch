package nsq_options

import "time"

// Options options for config
type Options struct {
	Topics               []string      `flag:"topic"`
	TopicPatterns        []string      `flag:"topic-pattern"`
	TopicRefreshInterval time.Duration `flag:"topic-refresh"`
	Channel              string        `flag:"channel"`

	NSQDTCPAddrs             []string      `flag:"nsqd-tcp-address"`
	NSQLookupdHTTPAddrs      []string      `flag:"lookupd-http-address"`
	ConsumerOpts             []string      `flag:"consumer-opt"`
	MaxInFlight              int           `flag:"max-in-flight"`
	HTTPClientConnectTimeout time.Duration `flag:"http-client-connect-timeout"`
	HTTPClientRequestTimeout time.Duration `flag:"http-client-request-timeout"`

	LogPrefix string `flag:"log-prefix"`
	LogLevel  string `flag:"log-level"`
	OutputDir string `flag:"output-dir"`
	WorkDir   string `flag:"work-dir"`
	// DatetimeFormat string        `flag:"datetime-format"`
	SyncInterval time.Duration `flag:"sync-interval"`
}

// NewOptions make Options
func NewOptions() *Options {
	return &Options{
		LogPrefix:                "[nsqToElasticsearch] ",
		LogLevel:                 "info",
		Channel:                  "nsqToElasticsearch",
		MaxInFlight:              200,
		OutputDir:                "/tmp",
		TopicRefreshInterval:     time.Minute,
		SyncInterval:             30 * time.Second,
		HTTPClientConnectTimeout: 2 * time.Second,
		HTTPClientRequestTimeout: 5 * time.Second,
	}
}
