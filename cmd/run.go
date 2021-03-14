package cmd

import (
	"fmt"
	"flag"
	"strings"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/nsqio/go-nsq"

	"github.com/JieTrancender/nsq_to_elasticsearch/internal/topic_discoverer"
	"github.com/JieTrancender/nsq_to_elasticsearch/internal/nsq_options"
)

const (
	DefaultConfigFilename = "nsqConsumer"

	EnvPrefix = "NsqConsumer"

	Version = "0.0.1"
)

type NsqConsumerConfig struct {
	EtcdEndpoint []string `json:"etcdEndpoint"`
	EtcdPath     string `json:"etcdPath"`
	EtcdUsername string `json:"etcdUsername"`
	EtcdPassword string `json:"etcdPassword"`
}

func genRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "start nsq consumer tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			flag.Parse()
			
			fmt.Printf("version: v%s\n", Version)
			// fmt.Printf("etcd config %s %s %s %s\n", etcdEndpoint, etcdPath, etcdUsername, etcdPassword)

			hupChan := make(chan os.Signal, 1)
			termChan := make(chan os.Signal, 1)
			signal.Notify(hupChan, syscall.SIGHUP)
			signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
			
			opts := nsq_options.NewOptions()
			cfg := nsq.NewConfig()
			
			ddAccessToken := "123456"
			discoverer, _ := topic_discoverer.NewTopicDiscoverer(opts, cfg, hupChan, termChan,
				ddAccessToken)
			discoverer.Run()
		},
	}

	return runCmd
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigName(DefaultConfigFilename)

	v.AddConfigPath("./config/")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.SetEnvPrefix(EnvPrefix)

	v.AutomaticEnv()

	bindFlags(cmd, v)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			_ = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", EnvPrefix, envVarSuffix))
		}

		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
