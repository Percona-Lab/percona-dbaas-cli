#!/bin/bash

set -o xtrace

S3_BUCKET="operator-testing"
S3_SECRET="minio-secret"
S3_ENDPOINT="http://minio-service:9000/"
S3_REGION="us-east-1"



bucket_name=$(kubectl get psmdb test-cluster -o jsonpath='{.spec.backup.storages.defaultS3Storage.s3.bucket}')
secret=$(kubectl get psmdb test-cluster -o jsonpath='{.spec.backup.storages.defaultS3Storage.s3.credentialsSecret}')
endpoint=$(kubectl get psmdb test-cluster -o jsonpath='{.spec.backup.storages.defaultS3Storage.s3.endpointUrl}')
region=$(kubectl get psmdb test-cluster -o jsonpath='{.spec.backup.storages.defaultS3Storage.s3.region}')

# if [[ ${bucket_name} != ${S3_BUCKET} || ${secret} != ${S3_SECRET} \
# 	  || ${endpoint} != ${S3_ENDPOINT} || ${region} != ${S3_REGION} ]]; then
# 	echo "S3 bucket settings has not been provisioned"
# 	exit 1
# fi

if [[ ${endpoint} != ${S3_ENDPOINT} || ${region} != ${S3_REGION} ]]; then
	echo "S3 bucket settings has not been provisioned"
	exit 1
fi