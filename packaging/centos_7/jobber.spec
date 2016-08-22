Name:		jobber
Version:	%{_pkg_version}
Release:	%{_pkg_release}%{?dist}
Summary:	A replacement for cron.

Group:		System Environment/Daemons
License:	MIT
URL:		http://dshearer.github.io/jobber/
Source0:	jobber-%{_pkg_version}.tgz
Source1:    jobber.service

%{?systemd_requires}
BuildRequires:	golang, coreutils, systemd

Prefix:         /usr/local

%define debug_package %{nil}

%description
A replacement for cron, with sophisticated status-reporting and error-handling.


%files
%attr(4755,jobber_client,root) /usr/local/bin/jobber
%attr(0755,root,root) /usr/local/libexec/jobberd
%attr(0644,root,root) %{_unitdir}/jobber.service


%prep

# move sources into BUILD
%setup -q
cp "%{_sourcedir}/jobber.service" "%{_builddir}/"

# create Go workspace
GO_WKSPC="%{_builddir}/go_workspace"
GO_SRC_DIR="${GO_WKSPC}/src/github.com/dshearer"
mkdir -p "${GO_SRC_DIR}"
ln -fs "%{_builddir}/jobber-%{_pkg_version}" "${GO_SRC_DIR}/jobber"

echo "GO_WKSPC=${GO_WKSPC}" > "%{_builddir}/vars"
echo "GO_SRC_DIR=${GO_SRC_DIR}" >> "%{_builddir}/vars"


%build
source "%{_builddir}/vars"
export GO_WKSPC
export GOPATH="${GO_WKSPC}"
make %{?_smp_mflags} -C "${GO_SRC_DIR}/jobber" check


%install
source "%{_builddir}/vars"
export GO_WKSPC
export GOPATH="${GO_WKSPC}"
%make_install -C "${GO_SRC_DIR}/jobber"
mkdir -p "%{buildroot}/%{_unitdir}"
cp "%{_builddir}/jobber.service" "%{buildroot}/%{_unitdir}/"


%pre
if [ "$1" -gt 1 ]; then
    userdel jobber_client 2>/dev/null ||:
fi
useradd --home / -M --system --shell /sbin/nologin jobber_client


%post
%systemd_post jobber.service
systemctl enable jobber.service


%preun
%systemd_preun jobber.service


%postun
%systemd_postun_with_restart jobber.service


%changelog