package cmd

import (
	"fmt"
	"flag"
	"strings"
	"os"
	"os/signal"
	"syscall"
	"reflect"

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
	etcdEndpoints := []string{}
	etcdUsername := ""
	etcdPassword := ""
	etcdPath := ""
	etcd_hosts := []string{}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "start nsq consumer tool",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			flag.Parse()
			
			fmt.Printf("version: v%s\n", Version)
			fmt.Printf("etcd config %v %s %s %s %v %v %v\n", etcdEndpoints, etcdPath, etcdUsername, etcdPassword, etcd_hosts, reflect.TypeOf(etcdEndpoints), reflect.TypeOf(etcd_hosts))

			hupChan := make(chan os.Signal, 1)
			termChan := make(chan os.Signal, 1)
			signal.Notify(hupChan, syscall.SIGHUP)
			signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
			
			opts := nsq_options.NewOptions()
			cfg := nsq.NewConfig()
			
			ddAccessToken := "123456"
			discoverer, err := topic_discoverer.NewTopicDiscoverer(opts, cfg, hupChan, termChan,
				ddAccessToken, etcdEndpoints, etcdUsername, etcdPassword, etcdPath)
			if err != nil {
				panic(err)
			}
			discoverer.Run()
		},
	}

	runCmd.Flags().StringArrayVar(&etcdEndpoints, "etcd_endpoints", []string{"127.0.0.1:2379"}, "etcd host(may be given multi times)")
	runCmd.Flags().StringVar(&etcdUsername, "etcd_username", "root", "username of etcd")
	runCmd.Flags().StringVar(&etcdPassword, "etcd_password", "123456", "password of etcd")
	runCmd.Flags().StringVar(&etcdPath, "etcd_path", "config/nsq_to_elasticsearch/default", "path of the config in etcd")

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

	// bindFlags(cmd, v)

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
