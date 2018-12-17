package db_broker

const (
	// Key to set instance id
	InstanceKey      = "instance"

	// Key to provision info
	ProvisionInfoKey = "provision-info"

	// The file path for checking the namespace in which the broker server is running
	NamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	// Name of the catelog
	CatelogKeyKubeDB = "kubedb"

	// Name of the providers
	ProviderNameMySQL         = "mysql"
	ProviderNamePostgreSQL    = "postgresql"
	ProviderNameElasticsearch = "elasticsearch"
	ProviderNameMongoDB       = "mongodb"
	ProviderNameRedis         = "redis"
	ProviderNameMemcached     = "memcached"

	// Versions used in demo plans of different databases
	demoMySQLVersion = "8.0-v1"
	demoPostgresVersion = "10.2-v1"
	demoElasticSearchVersion = "6.3-v1"
	demoMongoDBVersion = "3.6-v1"
	demoRedisVersion = "4.0-v1"
	demoMemcachedVersion = "1.5.4-v1"

	// Name of the plans of under different services
	planMySQLDemo = "demo-mysql"
	planMySQL = "mysql"

	planPostgresDemo = "demo-postgresql"
	planPostgresHADemo = "demo-ha-postgresql"
	planPostgres = "postgresql"

	planElasticSearchDemo = "demo-elasticsearch"
	planElasticSearchClusterDemo = "demo-elasticsearch-cluster"
	planElasticSearch = "elasticsearch"

	planMongoDBDemo = "demo-mongodb"
	planMongoDBClusterDemo = "demo-mongodb-cluster"
	planMongoDB = "mongodb"

	planRedisDemo = "demo-redis"
	planRedis = "redis"

	planMemcachedDemo = "demo-memcached"
	planMemcached = "memcached"
)
