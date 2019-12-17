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


%description
percona-dbaas-cli


%prep
%setup -q -n %{name}-%{version}


%build


%install
rm -rf $RPM_BUILD_ROOT
install -m 0755 -d $RPM_BUILD_ROOT/%{_bindir}
install -D -m 0755 linux/percona-dbaas $RPM_BUILD_ROOT/%{_bindir}/
install -D -m 0755 linux/percona-dbaas $RPM_BUILD_ROOT/%{_bindir}/kubectl-pxc
install -D -m 0755 linux/percona-dbaas $RPM_BUILD_ROOT/%{_bindir}/kubectl-psmdb
install -D -m 0755 linux/percona-kubectl $RPM_BUILD_ROOT/%{_bindir}/


%files
%defattr(-, root, root, -)
%license LICENSE
%{_bindir}/percona-dbaas
%{_bindir}/kubectl-pxc
%{_bindir}/kubectl-psmdb
%{_bindir}/percona-kubectl


%changelog
* Tue Dec 24 2019 Viacheslav Sarzhan <slava.sarzhan@percona.com>
- First build
