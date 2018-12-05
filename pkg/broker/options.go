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
	KubeConfig   string
	Async        bool
	StorageClass string
}

func NewExtraOptions() *ExtraOptions {
	return &ExtraOptions{
		CatalogPath:  "/etc/config/catalogs",
		Async:        false,
		StorageClass: "standard",
	}
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func (s *ExtraOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.CatalogPath, "catalog-path", s.CatalogPath, "The path to the catalog.")
	fs.StringSliceVar(&s.CatalogNames, "catalog-names", s.CatalogNames,
		"List of catalogs those can be run by this service-broker, comma separated.")
	fs.StringVar(&s.KubeConfig, "kube-config", s.KubeConfig, "specify the kube config path to be used.")
	fs.BoolVar(&s.Async, "async", s.Async, "Indicates whether the broker is handling the requests asynchronously.")
	fs.StringVar(&s.StorageClass, "storage-class", s.StorageClass,
		"name of the storage-class for database storage.")
}
