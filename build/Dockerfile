FROM registry.access.redhat.com/ubi7/ubi-minimal
RUN microdnf update -y && microdnf clean all

LABEL name="Percona Database As a Service Command Line Interface" \
      vendor="Percona" \
      summary="Percona DBAAS CLI management tool for PXC/PSMDB clusters administration" \
      description="Percona DBAAS CLI management tool for PXC/PSMDB clusters administration." \
      maintainer="Percona Development <info@percona.com>"

RUN rpmkeys --import https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg https://packages.cloud.google.com/yum/doc/yum-key.gpg

COPY build/google-cloud-sdk.repo /etc/yum.repos.d/
RUN microdnf install -y google-cloud-sdk kubectl

COPY LICENSE /licenses/
COPY build/_output/bin/percona-dbaas /usr/local/bin/percona-dbaas


USER nobody
