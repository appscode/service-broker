package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/kube-mon/api"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeEtcd     = "etc"
	ResourceKindEtcd     = "Etcd"
	ResourceSingularEtcd = "etcd"
	ResourcePluralEtcd   = "etcds"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Etcd defines a Etcd database.
type Etcd struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              EtcdSpec   `json:"spec,omitempty"`
	Status            EtcdStatus `json:"status,omitempty"`
}

type EtcdSpec struct {
	// Version of Etcd to be deployed.
	Version types.StrYo `json:"version"`
	// Number of instances to deploy for a Etcd database.
	Replicas *int32 `json:"replicas,omitempty"`
	// Storage spec to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`
	// Database authentication secret
	DatabaseSecret *core.SecretVolumeSource `json:"databaseSecret,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Init is used to initialize database
	// +optional
	Init *InitSpec `json:"init,omitempty"`
	// BackupSchedule spec to specify how database backup will be taken
	// +optional
	BackupSchedule *BackupScheduleSpec `json:"backupSchedule,omitempty"`
	// If DoNotPause is true, controller will prevent to delete this Postgres object.
	// Controller will create same Postgres object and ignore other process.
	// +optional
	DoNotPause bool `json:"doNotPause,omitempty"`
	// Monitor is used monitor database instance
	// +optional
	Monitor *api.AgentSpec `json:"monitor,omitempty"`
	// Compute Resources required by the sidecar container.
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *core.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []core.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// +optional
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type EtcdStatus struct {
	CreationTime *metav1.Time  `json:"creationTime,omitempty"`
	Phase        DatabasePhase `json:"phase,omitempty"`
	Reason       string        `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EtcdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of Etcd TPR objects
	Items []Etcd `json:"items,omitempty"`
}
