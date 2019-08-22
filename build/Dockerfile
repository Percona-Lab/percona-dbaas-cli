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
RUN curl -o /licenses/LICENSE.kubectl https://raw.githubusercontent.com/kubernetes/kubectl/master/LICENSE
RUN cp /usr/share/google-cloud-sdk/LICENSE /licenses/LICENSE.gcloud
COPY build/_output/bin/percona-dbaas /usr/local/bin/percona-dbaas

RUN install -d -o 99 -g 0 -m 0775 /tmp/dbaas
RUN sed -i "s/^nobody.*/nobody:x:99:99:Nobody:\/tmp\/dbaas:\/sbin\/nologin/g" /etc/passwd

EXPOSE 8081
USER nobody

CMD percona-dbaas service-broker
