package e2e

import (
	"flag"

	"github.com/appscode/go/flags"
	logs "github.com/appscode/go/log/golog"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

type E2EOptions struct {
	KubeContext string
	KubeConfig  string
}

var (
	options = &E2EOptions{
		KubeConfig: filepath.Join(homedir.HomeDir(), ".kube", "config"),
	}
	//brokerImageFlag = "shudipta/db-broker:try"
	//brokerImageFlag = "shudipta/db-broker:try-for-pgsql"
	//brokerImageFlag = "shudipta/db-broker:try-for-elasticsearch"
	//brokerImageFlag = "shudipta/db-broker:try-for-mongodb"
	brokerImageFlag = "shudipta/db-broker:try-for-redis"
)

func init() {
	flag.StringVar(&options.KubeConfig, "kubeconfig", options.KubeConfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&options.KubeContext, "kube-context", "", "Name of kube context")

	flag.StringVar(&brokerImageFlag, "broker-image", brokerImageFlag,
		"The container image for the broker to test against")
	//framework.RegisterParseFlags()

	enableLogging()
	flag.Parse()
}

func enableLogging() {
	defer func() {
		logs.InitLogs()
		defer logs.FlushLogs()
	}()
	flag.Set("logtostderr", "true")
	logLevelFlag := flag.Lookup("v")
	if logLevelFlag != nil {
		if len(logLevelFlag.Value.String()) > 0 && logLevelFlag.Value.String() != "0" {
			return
		}
	}
	flags.SetLogLevel(2)
}
