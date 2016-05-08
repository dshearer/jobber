Name:		jobber
Version:	%{_pkg_version}
Release:	%{_pkg_release}
Summary:	A replacement for cron.

Group:		System Environment/Daemons
License:	MIT
URL:		http://dshearer.github.io/jobber/
Source0:	../SOURCES/jobber-%{_pkg_version}.tgz
Source1:    ../SOURCES/jobber_init

BuildRequires:	golang, coreutils
Requires:	daemonize, initscripts

Prefix:         /usr/local

%define debug_package %{nil}

%description


%prep

# move sources into BUILD
%setup -q
cp "%{_sourcedir}/jobber_init" "%{_builddir}/"

# create Go workspace
GOPATH="%{_builddir}/go_workspace"
GO_SRC_DIR="${GOPATH}/src/github.com/dshearer"
mkdir -p "${GO_SRC_DIR}"
ln -fs "%{_builddir}/jobber-%{_pkg_version}" "${GO_SRC_DIR}/jobber"

echo "GOPATH=${GOPATH}" > "%{_builddir}/vars"
echo "GO_SRC_DIR=${GO_SRC_DIR}" >> "%{_builddir}/vars"


%build
#make %{?_smp_mflags}


%install
source "%{_builddir}/vars"
%make_install -C "${GO_SRC_DIR}/jobber"
mkdir -p "%{buildroot}/etc/init.d"
cp "%{_builddir}/jobber_init" "%{buildroot}/etc/init.d/jobber"


%files
%attr(4755,jobber_client,root) /usr/local/bin/jobber
%attr(0755,root,root) /usr/local/libexec/jobberd
%attr(0755,root,root) /etc/init.d/jobber

%pre
if [ "$1" -gt 1 ]; then
    userdel jobber_client 2>/dev/null ||:
fi
useradd --home / -M --system --shell /sbin/nologin jobber_client

%post
if [ "$1" -eq 1 ]; then
    service jobber start
else
    service jobber condrestart
fi

%preun
if [ "$1" -eq 0 ]; then
    service jobber stop 2>/dev/null ||:
fi

%postun
if [ "$1" -eq 0 ]; then
    userdel jobber_client
fi

%changelog

