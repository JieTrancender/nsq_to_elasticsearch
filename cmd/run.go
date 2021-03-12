package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DefaultConfigFilename = "nsqConsumer"

	EnvPrefix = "NsqConsumer"

	Version = "0.0.1"
)

type NsqConsumerConfig struct {
	EtcdEndpoint string `json:"etcdEndpoint"`
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
			fmt.Printf("version: v%s\n", Version)
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