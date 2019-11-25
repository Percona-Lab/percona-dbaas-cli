// Copyright © 2019 Percona, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type BackupState string

const (
	BackupUnknown   BackupState = "Unknown"
	BackupStarting              = "Starting"
	BackupRunning               = "Running"
	BackupFailed                = "Failed"
	BackupSucceeded             = "Succeeded"
)

const (
	DefaultBcpStorageName = "defaultS3Storage"
)

type BackupStorageType string

const (
	BackupStorageFilesystem BackupStorageType = "filesystem"
	BackupStorageS3         BackupStorageType = "s3"
)

type BackupStorageS3Spec struct {
	Bucket            string `json:"bucket"`
	CredentialsSecret string `json:"credentialsSecret"`
	Region            string `json:"region,omitempty"`
	EndpointURL       string `json:"endpointUrl,omitempty"`
}

type BackupStorageSpec struct {
	Type   BackupStorageType   `json:"type"`
	S3     BackupStorageS3Spec `json:"s3,omitempty"`
	Volume *VolumeSpec         `json:"volume,omitempty"`
}

type VolumeSpec struct {
	// EmptyDir to use as data volume for mysql. EmptyDir represents a temporary
	// directory that shares a pod's lifetime.
	// +optional
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`

	// HostPath to use as data volume for mysql. HostPath represents a
	// pre-existing file or directory on the host machine that is directly
	// exposed to the container.
	// +optional
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`

	// PersistentVolumeClaim to specify PVC spec for the volume for mysql data.
	// It has the highest level of precedence, followed by HostPath and
	// EmptyDir. And represents the PVC specification.
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

type S3StorageConfig struct {
	EndpointURL       string
	Bucket            string
	Region            string
	CredentialsSecret string
	KeyID             string
	Key               string
	SkipStorage       bool
}

type ErrNoS3Options string

func (e ErrNoS3Options) Error() string {
	return "not enough options to set S3 backup storage: " + string(e)
}

func (p Cmd) S3Storage(appName string, c S3StorageConfig /*f *pflag.FlagSet*/) (*BackupStorageSpec, error) {
	bucket := c.Bucket
	if bucket == "" {
		return nil, ErrNoS3Options("no bucket defined")
	}

	region := c.Region
	endpoint := c.EndpointURL

	secretName := c.CredentialsSecret
	if secretName == "" {
		keyid := c.KeyID
		key := c.Key
		if key == "" || keyid == "" {
			return nil, ErrNoS3Options("neither s3-credentials-secret nor s3-access-key-id and s3-secret-access-key defined")
		}

		secretName = "s3-" + appName + "-" + GenRandString(5)
		secretData := map[string][]byte{
			"AWS_ACCESS_KEY_ID":     []byte(keyid),
			"AWS_SECRET_ACCESS_KEY": []byte(key),
		}
		err := p.CreateSecret(secretName, secretData)
		if err != nil {
			return nil, errors.Wrap(err, "create secret")
		}
	}

	s3 := &BackupStorageSpec{
		Type: BackupStorageS3,
		S3: BackupStorageS3Spec{
			Bucket:            bucket,
			Region:            region,
			EndpointURL:       endpoint,
			CredentialsSecret: secretName,
		},
	}

	return s3, nil
}
