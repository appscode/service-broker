package broker

import (
	"flag"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type Options struct {
	CatalogPath  string
	CatalogNames string
	KubeConfig   string
	Async        bool
	StorageClass string
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func AddFlags(o *Options) {
	flag.StringVar(&o.CatalogPath, "catalog-path", "/etc/config/catalogs", "The path to the catalog")
	flag.StringVar(&o.CatalogNames, "catalog-names", o.CatalogNames,
		"List of catalogs those can be run by this service-broker")
	flag.StringVar(&o.KubeConfig, "kube-config", "", "specify the kube config path to be used")
	flag.BoolVar(&o.Async, "async", false, "Indicates whether the broker is handling the requests asynchronously.")
	flag.StringVar(&o.StorageClass, "storage-class", "standard",
		"name of the storage-class for database storage")
}
