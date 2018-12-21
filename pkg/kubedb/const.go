package kubedb

const (
	// Key to set instance id
	InstanceKey = "servicecatalog.k8s.io/instance-id"

	// Key to provision info
	ProvisionInfoKey = "servicecatalog.k8s.io/provision-info"

	// The file path for checking the namespace in which the broker server is running
	NamespaceFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	// Name of the providers
	KubeDBServiceElasticsearch = "315fc21c-829e-4aa1-8c16-f7921c33550d"
	KubeDBServiceMemcached     = "d88856cb-fe3f-4473-ba8b-641480da810f"
	KubeDBServiceMongoDB       = "d690058d-666c-45d8-ba98-fcb9fb47742e"
	KubeDBServiceMySQL         = "938a70c5-f2bc-4658-82dd-566bed7797e9"
	KubeDBServicePostgreSQL    = "2010d83f-d908-4d9f-879c-ce8f5f527f2a"
	KubeDBServiceRedis         = "ccfd1c81-e59f-4875-a39f-75ba55320ce0"

	// Versions used in demo plans of different databases
	demoMySQLVersion         = "8.0-v1"
	demoPostgresVersion      = "10.2-v1"
	demoElasticSearchVersion = "6.3-v1"
	demoMongoDBVersion       = "3.6-v1"
	demoRedisVersion         = "4.0-v1"
	demoMemcachedVersion     = "1.5.4-v1"

	// Name of the plans of under different services
	planElasticSearchDemo        = "c4e99557-3a81-452e-b9cf-660f01c155c0"
	planElasticSearchClusterDemo = "2f05622b-724d-458f-abc8-f223b1afa0b9"
	planElasticSearch            = "6fa212e2-e043-4ae9-91c2-8e5c4403d894"

	planMemcachedDemo = "af1ce2dc-5734-4e41-aaa2-8aa6a58d688f"
	planMemcached     = "d40e49b2-f8fb-4d47-96d3-35089bd0942d"

	planMongoDBDemo        = "498c12a6-7a68-4983-807b-75737f99062a"
	planMongoDBClusterDemo = "6af19c54-7757-42e5-bb74-b8350037c4a2"
	planMongoDB            = "e8f87ba6-0711-42db-a663-a3c75b78a541"

	planMySQLDemo = "1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417"
	planMySQL     = "6ed1ab9e-a640-4f26-9328-423b2e3816d7"

	planPostgresDemo   = "c4bcf392-7ebb-4623-a79d-13d00d761d56"
	planPostgresHADemo = "41818203-0e2d-4d30-809f-a60c8c73dae8"
	planPostgres       = "13373a9b-d5f5-4d9a-88df-d696bbc19071"

	planRedisDemo = "4b6ad8a7-272e-4cfd-bb38-5b9d4bd3962f"
	planRedis     = "45716530-cadb-4247-b06a-24a34200d734"
)
