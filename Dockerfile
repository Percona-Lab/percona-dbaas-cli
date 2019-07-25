FROM golang:1.12-alpine

RUN apk update && apk add bash
# Install kubectl from Docker Hub.
COPY --from=lachlanevenson/k8s-kubectl:v1.10.3 /usr/local/bin/kubectl /usr/local/bin/kubectl

ADD . /go/src/github.com/Percona-Lab/percona-dbaas-cli
EXPOSE 8081
WORKDIR /go/src/github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas
RUN go install

ENTRYPOINT ["percona-dbaas", "pxc-broker"]