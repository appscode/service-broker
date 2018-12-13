package db_broker

const (
	InstanceKey      = "instance"
	ProvisionInfoKey = "provision-info"

	NamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	CatelogKeyKubeDB = "kubedb"

	ProviderNameMySQL         = "mysql"
	ProviderNamePostgreSQL    = "postgresql"
	ProviderNameElasticsearch = "elasticsearch"
	ProviderNameMongoDB       = "mongodb"
	ProviderNameRedis         = "redis"
	ProviderNameMemcached     = "memcached"
)
