package kubedb

import (
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

var (
	WaitForMySQLBeReady         = waitForMySQLBeReady
	WaitForPostgreSQLBeReady    = waitForPostgreSQLBeReady
	WaitForElasticsearchBeReady = waitForElasticsearchBeReady
	WaitForMongoDbBeReady       = waitForMongoDbBeReady
	WaitForRedisBeReady         = waitForRedisBeReady
	WaitForMemcachedBeReady     = waitForMemcachedBeReady
)

func waitForMemcachedBeReady(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		mc, err := extClient.Memcacheds(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return mc.Status.Phase == api.DatabasePhaseRunning, nil
	})
}

func waitForRedisBeReady(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		rd, err := extClient.Redises(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return rd.Status.Phase == api.DatabasePhaseRunning, nil
	})
}

func waitForMongoDbBeReady(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		mg, err := extClient.MongoDBs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return mg.Status.Phase == api.DatabasePhaseRunning, nil
	})
}

func waitForElasticsearchBeReady(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		es, err := extClient.Elasticsearches(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return es.Status.Phase == api.DatabasePhaseRunning, nil
	})
}

func waitForPostgreSQLBeReady(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		pgsql, err := extClient.Postgreses(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return pgsql.Status.Phase == api.DatabasePhaseRunning, nil
	})
}

func waitForMySQLBeReady(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		mysql, err := extClient.MySQLs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return mysql.Status.Phase == api.DatabasePhaseRunning, nil
	})
}

func patchRedis(extClient cs.KubedbV1alpha1Interface, rd *api.Redis, transform func(*api.Redis) *api.Redis) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchRedis(extClient, rd, transform); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchMemcached(extClient cs.KubedbV1alpha1Interface, mc *api.Memcached, transform func(*api.Memcached) *api.Memcached) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchMemcached(extClient, mc, transform); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchMongoDb(extClient cs.KubedbV1alpha1Interface, mg *api.MongoDB, transform func(*api.MongoDB) *api.MongoDB) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchMongoDB(extClient, mg, transform); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchElasticsearch(extClient cs.KubedbV1alpha1Interface, es *api.Elasticsearch, transform func(*api.Elasticsearch) *api.Elasticsearch) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchElasticsearch(extClient, es, transform); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchPostgreSQL(extClient cs.KubedbV1alpha1Interface, pgsql *api.Postgres, transform func(*api.Postgres) *api.Postgres) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchPostgres(extClient, pgsql, transform); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchMySQL(extClient cs.KubedbV1alpha1Interface, mysql *api.MySQL, transform func(*api.MySQL) *api.MySQL) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchMySQL(extClient, mysql, transform); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func deleteInBackground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func patchDormantDatabase(extClient cs.KubedbV1alpha1Interface, name, namespace string) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		dormantDatabase, err := extClient.DormantDatabases(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		_, _, err = util.PatchDormantDatabase(extClient, dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
			in.Spec.WipeOut = true
			return in
		})
		return true, nil
	})
}
