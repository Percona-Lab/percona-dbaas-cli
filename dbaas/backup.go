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

package dbaas

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type BackupState string

const (
	BackupUnknown   BackupState = "Unknown"
	BackupStarting              = "Starting"
	BackupRunning               = "Running"
	BackupFailed                = "Failed"
	BackupSucceeded             = "Succeeded"

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
	Type BackupStorageType   `json:"type"`
	S3   BackupStorageS3Spec `json:"s3,omitempty"`
}

type ErrNoS3Options string

func (e ErrNoS3Options) Error() string {
	return "not enough options to set S3 backup storage: " + string(e)
}

func S3Storage(app Deploy, f *pflag.FlagSet) (*BackupStorageSpec, error) {
	bucket, err := f.GetString("s3-bucket")
	if err != nil {
		return nil, errors.New("undefined `s3-bucket`")
	}
	if bucket == "" {
		return nil, ErrNoS3Options("no buket defined")
	}

	region, err := f.GetString("s3-region")
	if err != nil {
		return nil, errors.New("undefined `s3-region`")
	}

	endpoint, err := f.GetString("s3-endpoint-url")
	if err != nil {
		return nil, errors.New("undefined `s3-endpoint-url`")
	}

	secretName, err := f.GetString("s3-credentials-secret")
	if err != nil {
		return nil, errors.New("undefined `s3-credentials-secret`")
	}
	if secretName == "" {
		keyid, err := f.GetString("s3-key-id")
		if err != nil {
			return nil, errors.New("undefined `s3-key-id`")
		}
		key, err := f.GetString("s3-key")
		if err != nil {
			return nil, errors.New("undefined `s3-key`")
		}

		if key == "" || keyid == "" {
			return nil, ErrNoS3Options("neither s3-credentials-secret nor s3-key-id and s3-key defined")
		}

		secretName = "s3-" + app.Name() + "-" + GenRandString(5)
		secretData := map[string][]byte{
			"AWS_ACCESS_KEY_ID":     []byte(keyid),
			"AWS_SECRET_ACCESS_KEY": []byte(key),
		}
		err = CreateSecret(secretName, secretData)
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
