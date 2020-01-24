package v120

import (
	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v120 "github.com/percona/percona-xtradb-cluster-operator/v120/pkg/apis/pxc/v1"
)

// PerconaXtraDBCluster is the Schema for the perconaxtradbclusters API
type PerconaXtraDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   v120.PerconaXtraDBClusterSpec   `json:"spec,omitempty"`
	Status v120.PerconaXtraDBClusterStatus `json:"status,omitempty"`
}

var defaultAffinityTopologyKey = "kubernetes.io/hostname"

func (cr *PerconaXtraDBCluster) GetName() string {
	return cr.ObjectMeta.Name
}

func (cr *PerconaXtraDBCluster) SetLabels(labels map[string]string) {
	cr.ObjectMeta.Labels = labels
}

func (cr *PerconaXtraDBCluster) MarshalRequests() error {
	_, err := cr.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	return err
}

func (cr *PerconaXtraDBCluster) GetCR() (string, error) {
	b, err := json.Marshal(cr)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(b), nil
}

// Upgrade upgrades culster with given images
func (cr *PerconaXtraDBCluster) Upgrade(imgs map[string]string) {
	if img, ok := imgs["pxc"]; ok {
		cr.Spec.PXC.Image = img
	}
	if img, ok := imgs["proxysql"]; ok {
		cr.Spec.ProxySQL.Image = img
	}
	if img, ok := imgs["backup"]; ok {
		cr.Spec.Backup.Image = img
	}
}

func (cr *PerconaXtraDBCluster) SetName(name string) {
	cr.ObjectMeta.Name = name
}

func (cr *PerconaXtraDBCluster) SetUsersSecretName(name string) {
	cr.Spec.SecretsName = name + "-secrets"
}

func (cr *PerconaXtraDBCluster) GetOperatorImage() string {
	return "percona/percona-xtradb-cluster-operator:1.2.0"
}

func (cr *PerconaXtraDBCluster) SetDefaults() error {
	one := intstr.FromInt(1)

	cr.TypeMeta.APIVersion = "pxc.percona.com/v1"
	cr.TypeMeta.Kind = "PerconaXtraDBCluster"
	cr.ObjectMeta.Finalizers = []string{"delete-pxc-pods-in-order"}

	cr.Spec.PXC = &v120.PodSpec{}
	cr.Spec.PXC.Size = 3
	cr.Spec.PXC.Image = "percona/percona-xtradb-cluster-operator:1.2.0-pxc"
	cr.Spec.PXC.Affinity = &v120.PodAffinity{
		TopologyKey: &defaultAffinityTopologyKey,
	}
	cr.Spec.PXC.PodDisruptionBudget = &v120.PodDisruptionBudgetSpec{
		MaxUnavailable: &one,
	}
	volPXC, _ := resource.ParseQuantity("6G")
	cr.Spec.PXC.VolumeSpec = &v120.VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volPXC},
			},
		},
	}

	cr.Spec.ProxySQL = &v120.PodSpec{}
	cr.Spec.ProxySQL.Enabled = true
	cr.Spec.ProxySQL.Size = 1
	cr.Spec.ProxySQL.Image = "percona/percona-xtradb-cluster-operator:1.2.0-proxysql"
	cr.Spec.ProxySQL.Affinity = &v120.PodAffinity{
		TopologyKey: &defaultAffinityTopologyKey,
	}
	cr.Spec.ProxySQL.PodDisruptionBudget = &v120.PodDisruptionBudgetSpec{
		MaxUnavailable: &one,
	}
	volProxy, _ := resource.ParseQuantity("1G")
	cr.Spec.ProxySQL.VolumeSpec = &v120.VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volProxy},
			},
		},
	}

	cr.Spec.Backup = &v120.PXCScheduledBackup{
		Image: "percona/percona-xtradb-cluster-operator:1.2.0-backup",
	}
	return nil
}
