FROM registry.access.redhat.com/ubi7/ubi
RUN yum update -y && yum clean all

LABEL name="Percona Database As a Service Command Line Interface" \
      vendor="Percona" \
      summary="Percona DBAAS CLI management tool for PXC/PSMDB clusters administration" \
      description="Percona DBAAS CLI management tool for PXC/PSMDB clusters administration." \
      maintainer="Percona Development <info@percona.com>"

COPY LICENSE /licenses/
COPY build/_output/bin/percona-dbaas /usr/local/bin/percona-dbaas
COPY build/google-cloud-sdk.repo /etc/yum.repos.d/

RUN yum install -y google-cloud-sdk kubectl

USER nobody
