package main

import (
	"log"

	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/service-broker/pkg/cmds"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	if err := cmds.NewRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
