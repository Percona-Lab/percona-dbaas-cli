package v110

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/percona/percona-server-mongodb-operator/v110/pkg/apis/psmdb/v1"
	"github.com/pkg/errors"
)

// PerconaServerMongoDB is the Schema for the perconaservermongodbs API
type PerconaServerMongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   v1.PerconaServerMongoDBSpec   `json:"spec,omitempty"`
	Status v1.PerconaServerMongoDBStatus `json:"status,omitempty"`
}

func (cr *PerconaServerMongoDB) GetSpec() interface{} {
	rs := v1.ReplsetSpec{}
	cr.Spec.Replsets = []*v1.ReplsetSpec{&rs}
	return cr.Spec
}

func (cr *PerconaServerMongoDB) GetName() string {
	return cr.ObjectMeta.Name
}

func (cr *PerconaServerMongoDB) SetName(name string) {
	cr.ObjectMeta.Name = name
}

func (cr *PerconaServerMongoDB) SetUsersSecretName(name string) {
	cr.Spec.Secrets = &v1.SecretsSpec{
		Users: name + "-psmdb-users-secrets",
	}
}

func (cr *PerconaServerMongoDB) GetOperatorImage() string {
	return "percona/percona-server-mongodb-operator:1.1.0"
}

func (cr *PerconaServerMongoDB) SetLabels(labels map[string]string) {
	cr.ObjectMeta.Labels = labels
}

func (cr *PerconaServerMongoDB) MarshalRequests() error {
	if len(cr.Spec.Replsets) == 0 {
		return errors.New("no replsets")
	}
	_, err := cr.Spec.Replsets[0].VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	return err
}

func (cr *PerconaServerMongoDB) GetCR() (string, error) {
	b, err := json.Marshal(cr)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(b), nil
}

// Upgrade upgrades culster with given images
func (cr *PerconaServerMongoDB) Upgrade(imgs map[string]string) {
	if img, ok := imgs["psmdb"]; ok {
		cr.Spec.Image = img
	}
	if img, ok := imgs["backup"]; ok {
		cr.Spec.Backup.Image = img
	}
}

func (cr *PerconaServerMongoDB) SetDefaults() error {
	rsName := "rs0"
	rs := &v1.ReplsetSpec{
		Name: rsName,
	}

	volSizeFlag := "6G"
	volSize, err := resource.ParseQuantity(volSizeFlag)
	if err != nil {
		return errors.Wrap(err, "storage-size")
	}
	rs.VolumeSpec = &v1.VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volSize},
			},
		},
	}
	rs.Size = int32(3)
	rs.Resources = &v1.ResourcesSpec{
		Requests: &v1.ResourceSpecRequirements{
			CPU:    "600m",
			Memory: "1G",
		},
	}
	psmdbtpk := "kubernetes.io/hostname"
	rs.Affinity = &v1.PodAffinity{
		TopologyKey: &psmdbtpk,
	}
	cr.Spec.Replsets = []*v1.ReplsetSpec{
		rs,
	}
	cr.TypeMeta.APIVersion = "psmdb.percona.com/v1"
	cr.TypeMeta.Kind = "PerconaServerMongoDB"

	cr.Spec.Image = "percona/percona-server-mongodb-operator:1.1.0-mongod4.0"

	return nil
}
