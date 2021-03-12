package main

import (
	"os"

	"github.com/JieTrancender/nsq_to_elasticsearch/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
