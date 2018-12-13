package db_broker

const (
	InstanceKey      = "instance"
	ProvisionInfoKey = "provision-info"

	NamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	CatelogKeyKubeDB      = "kubedb"
	CatelogKeyVoyager     = "voyager"
	CatelogKeyStash       = "stash"
	CatelogKeyKubed       = "kubed"
	CatelogKeySwift       = "swift"
	CatelogKeySearchlight = "searchlight"
	CatelogKeyGuard       = "guard"

	ProviderNameMySQL         = "mysql"
	ProviderNamePostgreSQL    = "postgresql"
	ProviderNameElasticsearch = "elasticsearch"
	ProviderNameMongoDB       = "mongodb"
	ProviderNameRedis         = "redis"
	ProviderNameMemcached     = "memcached"
)
