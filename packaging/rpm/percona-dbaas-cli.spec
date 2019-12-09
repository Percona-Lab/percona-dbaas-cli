%undefine _missing_build_ids_terminate_build
%global debug_package %{nil}
%{!?with_systemd:%global systemd 0}
%{?el7:          %global systemd 1}
%{?el8:          %global systemd 1}


Name:  percona-dbaas-cli
Version: @@VERSION@@
Release: @@RELEASE@@%{?dist}
Summary: percona-dbaas-cli

Group:  Applications/Databases
License: ASL 2.0
URL:  https://github.com/percona/percona-backup-mongodb
Source0: %{name}-%{version}.tar.gz

BuildRequires: golang

%description
percona-dbaas-cli

%prep
%setup -q -n %{name}-%{version}


%build
cd ../
export PATH=/usr/local/go/bin:${PATH}
export GOROOT="/usr/local/go/"
export GOPATH=$(pwd)/
export PATH="/usr/local/go/bin:$PATH:$GOPATH"
export GOBINPATH="/usr/local/go/bin"
export GO_BUILD_LDFLAGS="-w -s -X main.version=@@VERSION@@ -X main.commit=@@REVISION@@"
mkdir -p src/github.com/Percona-Lab/
mkdir -p src/k8s.io
mv %{name}-%{version} src/github.com/Percona-Lab/%{name}
mv src/github.com/Percona-Lab/%{name}/kubernetes src/k8s.io/
ln -s src/github.com/Percona-Lab/%{name} %{name}-%{version}
cd src/
go build -o percona-dbaas github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/cmd
go build -o percona-kubectl k8s.io/kubernetes/cmd/kubectl
cd %{_builddir}


%install
rm -rf $RPM_BUILD_ROOT
install -m 0755 -d $RPM_BUILD_ROOT/%{_bindir}
install -D -m 0755 %{_builddir}/src/percona-dbaas $RPM_BUILD_ROOT/%{_bindir}/
install -D -m 0755 %{_builddir}/src/percona-dbaas $RPM_BUILD_ROOT/%{_bindir}/kubectl-pxc
install -D -m 0755 %{_builddir}/src/percona-dbaas $RPM_BUILD_ROOT/%{_bindir}/kubectl-psmdb
install -D -m 0755 %{_builddir}/src/percona-kubectl $RPM_BUILD_ROOT/%{_bindir}/

%files
%defattr(-, root, root, -)
%license LICENSE
%{_bindir}/percona-dbaas
%{_bindir}/kubectl-pxc
%{_bindir}/kubectl-psmdb
%{_bindir}/percona-kubectl

%changelog
* Mon Apr 15 2019 Evgeniy Patlan <evgeniy.patlan@percona.com>
- First build
