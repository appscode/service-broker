package main

import (
	"log"

	"github.com/appscode/service-broker/pkg/cmds"
	"kmodules.xyz/client-go/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	if err := cmds.NewRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}
