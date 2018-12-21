package framework

import (
	"flag"
	"os"

	"github.com/onsi/ginkgo/config"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	RecommendedConfigPathEnvVar = "SERVICECATALOGCONFIG"
)

type TestContextType struct {
	KubeHost              string
	KubeConfig            string
	KubeContext           string
	ServiceCatalogHost    string
	ServiceCatalogConfig  string
	ServiceCatalogContext string
}

var TestContext TestContextType

// Register flags common to all e2e test suites.
func RegisterCommonFlags() {
	// Turn on verbose by default to get spec names
	config.DefaultReporterConfig.Verbose = true

	// Turn on EmitSpecProgress to get spec progress (especially on interrupt)
	config.GinkgoConfig.EmitSpecProgress = true

	// Randomize specs as well as suites
	config.GinkgoConfig.RandomizeAllSpecs = true

	flag.StringVar(&TestContext.KubeHost, "kubernetes-host", "http://127.0.0.1:8080", "The kubernetes host, or apiserver, to connect to")
	flag.StringVar(&TestContext.KubeConfig, "kubernetes-config", os.Getenv(clientcmd.RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for kubernetes. Default value is from environment variable "+clientcmd.RecommendedConfigPathEnvVar)
	flag.StringVar(&TestContext.KubeContext, "kubernetes-context", "", "config context to use for kuberentes. If unset, will use value from 'current-context'")
	flag.StringVar(&TestContext.ServiceCatalogHost, "service-catalog-host", "http://127.0.0.1:30000", "The service catalog host, or apiserver, to connect to")
	flag.StringVar(&TestContext.ServiceCatalogConfig, "service-catalog-config", os.Getenv(RecommendedConfigPathEnvVar), "Path to config containing embedded authinfo for service catalog. Default value is from environment variable "+RecommendedConfigPathEnvVar)
	flag.StringVar(&TestContext.ServiceCatalogContext, "service-catalog-context", "", "config context to use for service catalog. If unset, will use value from 'current-context'")
}

func RegisterParseFlags() {
	RegisterCommonFlags()
	flag.Parse()
}
