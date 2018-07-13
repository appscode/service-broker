package e2e

import (
	"flag"
	//"github.com/kubedb/service-broker/test/e2e/framework"
	//"k8s.io/client-go/tools/clientcmd"
	//"github.com/golang/glog"
	logs "github.com/appscode/go/log/golog"
	"github.com/appscode/go/flags"
	"path/filepath"
	"k8s.io/client-go/util/homedir"
)

type E2EOptions struct {
	KubeContext        string
	KubeConfig         string
}

var (
	options = &E2EOptions{
		KubeConfig:         filepath.Join(homedir.HomeDir(), ".kube", "config"),
	}
)

func init() {
	flag.StringVar(&options.KubeConfig, "kubeconfig", options.KubeConfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&options.KubeContext, "kube-context", "", "Name of kube context")

	//flag.StringVar(&brokerImageFlag, "broker-image", "shudipta/servicebroker:try-as-minibroker",
	//	"The container image for the broker to test against")
	//framework.RegisterParseFlags()
	//
	//if "" == framework.TestContext.KubeConfig {
	//	glog.Fatalf("environment variable %v must be set", clientcmd.RecommendedConfigPathEnvVar)
	//}
	//if "" == framework.TestContext.ServiceCatalogConfig {
	//	glog.Fatalf("environment variable %v must be set", framework.RecommendedConfigPathEnvVar)
	//}
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