package broker

import (
	"github.com/spf13/pflag"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type ExtraOptions struct {
	CatalogPath  string
	CatalogNames []string
	Async        bool

	MasterURL      string
	KubeconfigPath string
	QPS            float64
	Burst          int
}

func NewExtraOptions() *ExtraOptions {
	return &ExtraOptions{
		CatalogPath: "/etc/config/catalogs",
		Async:       false,
		QPS:         100,
		Burst:       100,
	}
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func (s *ExtraOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.MasterURL, "master", s.MasterURL, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	fs.StringVar(&s.KubeconfigPath, "kubeconfig", s.KubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	fs.Float64Var(&s.QPS, "qps", s.QPS, "The maximum QPS to the master from this client")
	fs.IntVar(&s.Burst, "burst", s.Burst, "The maximum burst for throttle")

	fs.StringVar(&s.CatalogPath, "catalog-path", s.CatalogPath, "The path to the catalog.")
	fs.StringSliceVar(&s.CatalogNames, "catalog-names", s.CatalogNames,
		"List of catalogs those can be run by this service-broker, comma separated.")
	fs.BoolVar(&s.Async, "async", s.Async, "Indicates whether the broker is handling the requests asynchronously.")
}
