Name:		jobber
Version:	%{_pkg_version}
Release:	%{_pkg_release}
Summary:	A replacement for cron.

Group:		System Environment/Daemons
License:	MIT
URL:		http://dshearer.github.io/jobber/
Source0:	jobber-%{_pkg_version}.tgz
Source1:    jobber_init
Source2:    jobber.fc
Source3:    jobber.if
Source4:    jobber.te

BuildRequires:	golang >= 1.8, coreutils, rsync
Requires:	daemonize, initscripts

Prefix:         /usr/local

%define debug_package %{nil}

%description
A replacement for cron, with sophisticated status-reporting and error-handling.


%package selinux
Summary:        An SELinux policy for Jobber.
Group:          System Environment/Daemons
BuildRequires:  checkpolicy, selinux-policy-devel
%if "%{_selinux_policy_version}" != ""
Requires:       selinux-policy >= %{_selinux_policy_version}
%endif
Requires:       %{name} = %{version}-%{release}
Requires(post): /usr/sbin/semodule, /sbin/restorecon
Requires(postun): /usr/sbin/semodule, /sbin/restorecon
%description selinux
An SELinux policy for Jobber.


%files
%attr(0755,root,root) /usr/local/bin/jobber
%attr(0755,root,root) /usr/local/libexec/jobbermaster
%attr(0755,root,root) /usr/local/libexec/jobberrunner
%attr(0755,root,root) /etc/init.d/jobber
%config(noreplace)    /etc/jobber.conf

%files selinux
%defattr(-,root,root,0755)
%doc selinux/*
%{_datadir}/selinux/*/jobber.pp


%prep

# move sources into BUILD
%setup -q
cp "%{_sourcedir}/jobber_init" "%{_builddir}/"

# create Go workspace
GO_WKSPC="%{_builddir}/go_workspace"
GO_SRC_DIR="${GO_WKSPC}/src/github.com/dshearer"
mkdir -p "${GO_SRC_DIR}"
ln -fs "%{_builddir}/jobber-%{_pkg_version}" "${GO_SRC_DIR}/jobber"

echo "GO_WKSPC=${GO_WKSPC}" > "%{_builddir}/vars"
echo "GO_SRC_DIR=${GO_SRC_DIR}" >> "%{_builddir}/vars"

# SELinux stuff
mkdir -p selinux
cp -p %{SOURCE2} %{SOURCE3} %{SOURCE4} selinux/


%build
source "%{_builddir}/vars"
export GO_WKSPC
export GOPATH="${GO_WKSPC}"
make %{?_smp_mflags} -C "${GO_SRC_DIR}/jobber" check

# SELinux stuff
cd selinux
for selinuxvariant in %{selinux_variants}
do
  make NAME=${selinuxvariant} -f /usr/share/selinux/devel/Makefile
  mv jobber.pp jobber.pp.${selinuxvariant}
  make NAME=${selinuxvariant} -f /usr/share/selinux/devel/Makefile clean
done
cd -


%install
source "%{_builddir}/vars"
export GO_WKSPC
export GOPATH="${GO_WKSPC}"
%make_install -C "${GO_SRC_DIR}/jobber"
mkdir -p "%{buildroot}/etc/init.d"
cp "%{_builddir}/jobber_init" "%{buildroot}/etc/init.d/jobber"

# SELinux stuff
for selinuxvariant in %{selinux_variants}
do
  install -d %{buildroot}%{_datadir}/selinux/${selinuxvariant}
  install -p -m 644 selinux/jobber.pp.${selinuxvariant} \
    %{buildroot}%{_datadir}/selinux/${selinuxvariant}/jobber.pp
done


%post
if [ "$1" -eq 1 ]; then
    /sbin/service jobber start
    /sbin/chkconfig --add jobber
    /sbin/chkconfig jobber on
else
    /sbin/service jobber condrestart
fi


%preun
if [ "$1" -eq 0 ]; then
    /sbin/service jobber stop 2>/dev/null ||:
fi


%postun
if [ "$1" -eq 0 ]; then
    /sbin/chkconfig jobber off
    /sbin/chkconfig --del jobber
fi


%post selinux
for selinuxvariant in %{selinux_variants}
do
  /usr/sbin/semodule -s ${selinuxvariant} -i \
    %{_datadir}/selinux/${selinuxvariant}/jobber.pp &> /dev/null ||:
done
restorecon -Rv /usr/local /etc/init.d ||:
/sbin/service jobber condrestart


%postun selinux
RUNNING=false
if /sbin/service jobber status; then
    RUNNING=true
    /sbin/service jobber stop
fi
if [ $1 -eq 0 ] ; then
  for selinuxvariant in %{selinux_variants}
  do
    /usr/sbin/semodule -s ${selinuxvariant} -r jobber &> /dev/null ||:
  done
fi
if [ "${RUNNING}" = "true" ]; then
    /sbin/service jobber start
fi


%changelog