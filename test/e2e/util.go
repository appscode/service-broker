package e2e

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
