package cmd

import (
	"github.com/spf13/cobra"
)

type NsqConsumerCmd struct {
	cobra.Command
	RunCmd *cobra.Command
}

// NewRootCmd creates root command
func NewRootCmd() *NsqConsumerCmd {
	rootCmd := &NsqConsumerCmd{}

	rootCmd.RunCmd = genRunCmd()

	rootCmd.AddCommand(rootCmd.RunCmd)

	return rootCmd
}
