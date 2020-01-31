package v130

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v130 "github.com/percona/percona-server-mongodb-operator/v130/pkg/apis/psmdb/v1"
	"github.com/pkg/errors"
)

// PerconaServerMongoDB is the Schema for the perconaservermongodbs API
type PerconaServerMongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   v130.PerconaServerMongoDBSpec   `json:"spec,omitempty"`
	Status v130.PerconaServerMongoDBStatus `json:"status,omitempty"`
}

func (cr *PerconaServerMongoDB) GetSpec() interface{} {
	rs := v130.ReplsetSpec{}
	cr.Spec.Replsets = []*v130.ReplsetSpec{&rs}
	return cr.Spec
}

func (cr *PerconaServerMongoDB) GetName() string {
	return cr.ObjectMeta.Name
}

func (cr *PerconaServerMongoDB) SetName(name string) {
	cr.ObjectMeta.Name = name
}

func (cr *PerconaServerMongoDB) SetUsersSecretName(name string) {
	cr.Spec.Secrets = &v130.SecretsSpec{
		Users: name + "-psmdb-users-secrets",
	}
}

func (cr *PerconaServerMongoDB) GetOperatorImage() string {
	return "percona/percona-server-mongodb-operator:1.3.0"
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
	rs := &v130.ReplsetSpec{
		Name: rsName,
	}

	volSizeFlag := "6G"
	volSize, err := resource.ParseQuantity(volSizeFlag)
	if err != nil {
		return errors.Wrap(err, "storage-size")
	}
	rs.VolumeSpec = &v130.VolumeSpec{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: volSize},
			},
		},
	}
	rs.Size = int32(3)
	rs.Resources = &v130.ResourcesSpec{
		Requests: &v130.ResourceSpecRequirements{
			CPU:    "600m",
			Memory: "1G",
		},
	}
	psmdbtpk := "none" //"kubernetes.io/hostname"
	rs.Affinity = &v130.PodAffinity{
		TopologyKey: &psmdbtpk,
	}
	cr.Spec.Replsets = []*v130.ReplsetSpec{
		rs,
	}
	cr.TypeMeta.APIVersion = "psmdb.percona.com/v1-3-0"
	cr.TypeMeta.Kind = "PerconaServerMongoDB"

	cr.Spec.Image = "percona/percona-server-mongodb-operator:1.3.0-mongod4.0"

	f := false
	op := v130.MongodSpecOperationProfiling{
		Mode:      "all",
		RateLimit: 1,
	}
	sec := v130.MongodSpecSecurity{
		EnableEncryption: &f,
	}
	mongod := v130.MongodSpec{
		OperationProfiling: &op,
		Security:           &sec,
	}
	cr.Spec.Mongod = &mongod
	cr.Spec.PMM.Enabled = false
	cr.Spec.PMM.ServerHost = "monitoring-service"
	cr.Spec.PMM.Image = "percona/percona-server-mongodb-operator:1.3.0-pmm"

	return nil
}
