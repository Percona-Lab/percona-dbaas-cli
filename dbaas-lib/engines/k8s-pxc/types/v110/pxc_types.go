package v110

import (
	"encoding/json"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/percona/percona-xtradb-cluster-operator/v110/pkg/apis/pxc/v1"
)

// PerconaXtraDBCluster is the Schema for the perconaxtradbclusters API
type PerconaXtraDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   v1.PerconaXtraDBClusterSpec   `json:"spec,omitempty"`
	Status v1.PerconaXtraDBClusterStatus `json:"status,omitempty"`
}

var defaultAffinityTopologyKey = "kubernetes.io/hostname"
var affinityTopologyKeyOff = "none"

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

// SetupMiniConfig is for seeting up config for working with minishift and minikube
func (cr *PerconaXtraDBCluster) SetupMiniConfig() {
	none := affinityTopologyKeyOff
	cr.Spec.PXC.Affinity.TopologyKey = &none
	cr.Spec.PXC.Resources = nil
	cr.Spec.ProxySQL.Affinity.TopologyKey = &none
	cr.Spec.ProxySQL.Resources = nil
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
	return "percona/percona-xtradb-cluster-operator:1.1.0"
}

func (cr *PerconaXtraDBCluster) GetProxysqlServiceType() string {
	if cr.Spec.ProxySQL != nil && cr.Spec.ProxySQL.ServiceType != nil {
		return string(*cr.Spec.ProxySQL.ServiceType)
	}

	return ""
}

func (cr *PerconaXtraDBCluster) GetStatus() string {
	return string(cr.Status.Status)
}

func (cr *PerconaXtraDBCluster) GetPXCStatus() string {
	return string(cr.Status.PXC.Status)
}

func (cr *PerconaXtraDBCluster) GetStatusHost() string {
	return cr.Status.Host
}

func (cr *PerconaXtraDBCluster) SetDefaults() error {
	one := intstr.FromInt(1)

	cr.TypeMeta.APIVersion = "pxc.percona.com/v1"
	cr.TypeMeta.Kind = "PerconaXtraDBCluster"
	cr.ObjectMeta.Finalizers = []string{"delete-pxc-pods-in-order"}

	cr.Spec.PXC = &v1.PodSpec{}
	cr.Spec.PXC.Size = 3
	cr.Spec.PXC.Image = "percona/percona-xtradb-cluster-operator:1.1.0-pxc"
	cr.Spec.PXC.Affinity = &v1.PodAffinity{
		TopologyKey: &defaultAffinityTopologyKey,
	}
	cr.Spec.PXC.PodDisruptionBudget = &v1.PodDisruptionBudgetSpec{
		MaxUnavailable: &one,
	}
	volPXC, _ := resource.ParseQuantity("6G")
	cr.Spec.PXC.VolumeSpec = &v1.VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volPXC},
			},
		},
	}

	cr.Spec.ProxySQL = &v1.PodSpec{}
	cr.Spec.ProxySQL.Enabled = true
	cr.Spec.ProxySQL.Size = 1
	cr.Spec.ProxySQL.Image = "percona/percona-xtradb-cluster-operator:1.1.0-proxysql"
	cr.Spec.ProxySQL.Affinity = &v1.PodAffinity{
		TopologyKey: &defaultAffinityTopologyKey,
	}
	cr.Spec.ProxySQL.PodDisruptionBudget = &v1.PodDisruptionBudgetSpec{
		MaxUnavailable: &one,
	}
	volProxy, _ := resource.ParseQuantity("1G")
	cr.Spec.ProxySQL.VolumeSpec = &v1.VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volProxy},
			},
		},
	}
	pmm := v1.PMMSpec{
		Enabled:    false,
		ServerHost: "monitoring-service",
		Image:      "percona/percona-xtradb-cluster-operator:1.1.0-pmm",
	}
	cr.Spec.PMM = &pmm

	cr.Spec.Backup = &v1.PXCScheduledBackup{
		Image: "percona/percona-xtradb-cluster-operator:1.1.0-backup",
	}
	return nil
}
