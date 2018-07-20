package db_broker

import (
	"github.com/appscode/kutil"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

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

func patchElasticsearch(extClient cs.KubedbV1alpha1Interface, es *api.Elasticsearch) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchElasticsearch(extClient, es, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.DoNotPause = false
			return in
		}); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchPostgreSQL(extClient cs.KubedbV1alpha1Interface, pgsql *api.Postgres) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchPostgres(extClient, pgsql, func(in *api.Postgres) *api.Postgres {
			in.Spec.DoNotPause = false
			return in
		}); err != nil {
			return false, nil
		}

		return true, nil
	})
}

func patchMySQL(extClient cs.KubedbV1alpha1Interface, mysql *api.MySQL) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		if _, _, err := util.PatchMySQL(extClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Spec.DoNotPause = false
			return in
		}); err != nil {
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
