package e2e

import (
	"github.com/appscode/go/types"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"path/filepath"
)

func Int32Ptr(i int32) *int32 {
	return &i
}

func newObjectMeta(name, namespace, label string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"app": label,
		},
	}
}

func NewCatalogConfigMap(name, namespace string) (*corev1.ConfigMap, error) {
	var (
		data          []byte
		err           error
		mysql         string
		postgresql    string
		elasticsearch string
		mongodb       string
		memcached     string
		redis         string
	)
	catalogPath := filepath.Join("..", "..", "hack", "deploy", "catalogs")
	if data, err = ioutil.ReadFile(filepath.Join(catalogPath, "mysql.yaml")); err != nil {
		return nil, err
	}
	mysql = string(data)
	if data, err = ioutil.ReadFile(filepath.Join(catalogPath, "postgresql.yaml")); err != nil {
		return nil, err
	}
	postgresql = string(data)
	if data, err = ioutil.ReadFile(filepath.Join(catalogPath, "elasticsearch.yaml")); err != nil {
		return nil, err
	}
	elasticsearch = string(data)
	if data, err = ioutil.ReadFile(filepath.Join(catalogPath, "mongodb.yaml")); err != nil {
		return nil, err
	}
	mongodb = string(data)
	if data, err = ioutil.ReadFile(filepath.Join(catalogPath, "memcached.yaml")); err != nil {
		return nil, err
	}
	memcached = string(data)
	if data, err = ioutil.ReadFile(filepath.Join(catalogPath, "redis.yaml")); err != nil {
		return nil, err
	}
	redis = string(data)

	return &corev1.ConfigMap{
		ObjectMeta: newObjectMeta(name, namespace, name),
		Data: map[string]string{
			"mysql.yaml":         mysql,
			"postgresql.yaml":    postgresql,
			"elasticsearch.yaml": elasticsearch,
			"mongodb.yaml":       mongodb,
			"memcached.yaml":     memcached,
			"redis.yaml":         redis,
		},
	}, nil
}

func NewServiceBrokerDeployment(name, namespace, image, storageClass string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: newObjectMeta(name, namespace, name),
		Spec: appsv1.DeploymentSpec{
			Replicas: Int32Ptr(1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: newObjectMeta("", "", name),
				Spec: corev1.PodSpec{
					ServiceAccountName: name,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"service-broker",
							},
							Args: []string{
								"--port",
								"8080",
								"--catalog-path",
								"/etc/config/catalogs",
								"-v",
								"5",
								"-logtostderr",
								"--storage-class",
								storageClass,
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/etc/config/catalogs",
									Name:      "catalogs-volume",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "catalogs-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: name,
									},
									DefaultMode: types.Int32P(511),
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewServiceBrokerService(name, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: newObjectMeta(name, namespace, name),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
}

func NewServiceBrokerServiceAccount(name, namspace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: newObjectMeta(name, namspace, name),
	}
}

func NewServiceBrokerClusterRoleBinding(name, namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: newObjectMeta(name, namespace, name),
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      name,
				Namespace: namespace,
			},
		},
	}
}
