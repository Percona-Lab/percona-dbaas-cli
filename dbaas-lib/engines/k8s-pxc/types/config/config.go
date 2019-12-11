package config

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
)

type ClusterConfig struct {
	OperatorImage string              `json:"operatorImage,omitempty"`
	Labels        map[string]string   `json:"labels,omitempty"`
	Name          string              `json:"name,omitempty"`
	SecretsName   string              `json:"secretsName,omitempty"`
	PXC           Spec                `json:"pxc,omitempty"`
	ProxySQL      Spec                `json:"proxySQL,omitempty"`
	S3            k8s.S3StorageConfig `json:"s3,omitempty"`
	PMM           PMMSpec             `json:"pmm,omitempty"`
	Backup        PXCScheduledBackup  `json:"backup,omitempty"`
}

type Spec struct {
	BrokerInstance string `json:"brokerInstance,omitempty"`
	StorageSize    string `json:"storageSize,omitempty"`
	//StorageClass   string `json:"storageClass,omitempty"`

	Enabled             bool                          `json:"enabled,omitempty"`
	Size                int32                         `json:"size,omitempty"`
	Image               string                        `json:"image,omitempty"`
	Resources           PodResources                  `json:"resources,omitempty"`
	VolumeSpec          VolumeSpec                    `json:"volumeSpec,omitempty"`
	Affinity            PodAffinity                   `json:"affinity,omitempty"`
	NodeSelector        map[string]string             `json:"nodeSelector,omitempty"`
	Tolerations         []corev1.Toleration           `json:"tolerations,omitempty"`
	PriorityClassName   string                        `json:"priorityClassName,omitempty"`
	Annotations         map[string]string             `json:"annotations,omitempty"`
	Labels              map[string]string             `json:"labels,omitempty"`
	ImagePullSecrets    []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	AllowUnsafeConfig   bool                          `json:"allowUnsafeConfigurations,omitempty"`
	Configuration       string                        `json:"configuration,omitempty"`
	PodDisruptionBudget PodDisruptionBudgetSpec       `json:"podDisruptionBudget,omitempty"`
}

type PodDisruptionBudgetSpec struct {
	MinAvailable   intstr.IntOrString `json:"minAvailable,omitempty"`
	MaxUnavailable intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

type PodAffinity struct {
	TopologyKey string          `json:"antiAffinityTopologyKey,omitempty"`
	Advanced    corev1.Affinity `json:"advanced,omitempty"`
}

type PodResources struct {
	Requests ResourcesList `json:"requests,omitempty"`
	Limits   ResourcesList `json:"limits,omitempty"`
}

type ResourcesList struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}

type VolumeSpec struct {
	EmptyDir              corev1.EmptyDirVolumeSource      `json:"emptyDir,omitempty"`
	HostPath              corev1.HostPathVolumeSource      `json:"hostPath,omitempty"`
	PersistentVolumeClaim corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

type PXCScheduledBackup struct {
	Image            string                           `json:"image,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference    `json:"imagePullSecrets,omitempty"`
	Schedule         []PXCScheduledBackupSchedule     `json:"schedule,omitempty"`
	Storages         map[string]k8s.BackupStorageSpec `json:"storages,omitempty"`
}

type PXCScheduledBackupSchedule struct {
	Name        string `json:"name,omitempty"`
	Schedule    string `json:"schedule,omitempty"`
	Keep        int    `json:"keep,omitempty"`
	StorageName string `json:"storageName,omitempty"`
}

type PMMSpec struct {
	Enabled    bool   `json:"enabled,omitempty"`
	ServerHost string `json:"serverHost,omitempty"`
	Image      string `json:"image,omitempty"`
	ServerUser string `json:"serverUser,omitempty"`
	ServerPass string `json:"serverPass,omitempty"`
}
