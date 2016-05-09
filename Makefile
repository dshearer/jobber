# GNU-standard vars (cf. http://www.gnu.org/prep/standards/html_node/Makefile-Conventions.html)
SHELL = /bin/sh
prefix = /usr/local
exec_prefix = ${prefix}
bindir = ${exec_prefix}/bin
libexecdir = ${exec_prefix}/libexec
srcdir = .
INSTALL = install
INSTALL_PROGRAM = ${INSTALL}

workdir = .

GOPATH ?= ${HOME}/go_workspace
CLIENT = jobber
DAEMON = jobberd
LIB = jobber.a
CLIENT_USER = jobber_client
TEST_TMPDIR = ${PWD}
JOBBER_PKG_NAME = jobber-$(shell cat ${srcdir}/version)

# read lists of source files
include sources.mk jobber/sources.mk jobberd/sources.mk
FINAL_LIB_SOURCES := ${LIB_SOURCES}
FINAL_LIB_TEST_SOURCES := ${LIB_TEST_SOURCES}
FINAL_CLIENT_SOURCES := $(CLIENT_SOURCES:%=jobber/%)
FINAL_CLIENT_TEST_SOURCES := $(CLIENT_TEST_SOURCES:%=jobber/%)
FINAL_DAEMON_SOURCES := $(DAEMON_SOURCES:%=jobberd/%)
FINAL_DAEMON_TEST_SOURCES := $(DAEMON_TEST_SOURCES:%=jobberd/%)

GO_SOURCES := \
	${FINAL_LIB_SOURCES} \
	${FINAL_CLIENT_SOURCES} \
	${FINAL_DAEMON_SOURCES}
OTHER_SOURCES := \
	Makefile \
	sources.mk \
	jobber/sources.mk \
	jobberd/sources.mk \
	buildtools \
	README.md \
	LICENSE \
	version \
	Godeps

ALL_SOURCES := \
	${GO_SOURCES} \
	${OTHER_SOURCES}

LDFLAGS = -ldflags "-X github.com/dshearer/jobber.jobberVersion=`cat version`"

SE_FILES = se_policy/jobber.fc \
           se_policy/jobber.if \
           se_policy/jobber.te \
           ${wildcard se_policy/include/**} \
           se_policy/Makefile \
           se_policy/policygentool

.PHONY : all
all : lib ${GOPATH}/bin/${CLIENT} ${GOPATH}/bin/${DAEMON}

.PHONY : check
check : ${FINAL_LIB_TEST_SOURCES} ${FINAL_CLIENT_TEST_SOURCES} ${FINAL_DAEMON_TEST_SOURCES}
	TMPDIR="${TEST_TMPDIR}" go test github.com/dshearer/jobber/jobberd

.PHONY : installcheck
installcheck :
	./test_installation

.PHONY : installdirs
installdirs :
	"${srcdir}/buildtools/mkinstalldirs" "${DESTDIR}${bindir}" "${DESTDIR}${libexecdir}"

.PHONY : install
install : installdirs ${GOPATH}/bin/${CLIENT} ${GOPATH}/bin/${DAEMON}
	# install files
	"${INSTALL_PROGRAM}" "${GOPATH}/bin/${CLIENT}" -t "${DESTDIR}${bindir}"
	"${INSTALL_PROGRAM}" "${GOPATH}/bin/${DAEMON}" -t "${DESTDIR}${libexecdir}"
	
	# change owner and perms
	-chown "${CLIENT_USER}:root" "${DESTDIR}${bindir}/${CLIENT}" && \
		chmod 4755 "${DESTDIR}${bindir}/${CLIENT}"
	-chown root:root "${DESTDIR}${libexecdir}/${DAEMON}" && \
		chmod 0755 "${DESTDIR}${libexecdir}/${DAEMON}"

.PHONY : uninstall
uninstall :
	-rm "${DESTDIR}${bindir}/${CLIENT}" "${DESTDIR}${bindir}/${DAEMON}"

dist : ${ALL_SOURCES}
	-mkdir -p "${workdir}/${JOBBER_PKG_NAME}" "${DESTDIR}"
	"${srcdir}/buildtools/srcsync" ${ALL_SOURCES} "${workdir}/${JOBBER_PKG_NAME}"
	tar -C "${workdir}" -czf "${DESTDIR}${JOBBER_PKG_NAME}.tgz" "${JOBBER_PKG_NAME}"
	rm -rf "${workdir}/${JOBBER_PKG_NAME}"

.PHONY : clean
clean :
	go clean -i github.com/dshearer/jobber
	go clean -i "github.com/dshearer/jobber/${CLIENT}"
	go clean -i "github.com/dshearer/jobber/${DAEMON}"
	rm -f "${DESTDIR}${JOBBER_PKG_NAME}.tgz"
	


.PHONY : lib
lib : ${FINAL_LIB_SOURCES}
	go install ${LDFLAGS} "github.com/dshearer/jobber"

${GOPATH}/bin/${CLIENT} : ${FINAL_CLIENT_SOURCES} lib
	go install ${LDFLAGS} "github.com/dshearer/jobber/${CLIENT}"

${GOPATH}/bin/${DAEMON} : ${FINAL_DAEMON_SOURCES} lib
	go install ${LDFLAGS} "github.com/dshearer/jobber/${DAEMON}"


## OLD:

/etc/init.d/jobber : jobber_init
	install -T -o root -g root -m 0755 "$<" "$@"
	chkconfig --add jobber
	chkconfig jobber on

/var/lock/subsys/jobber : ${DESTDIR}/sbin/${DAEMON} /etc/init.d/jobber
	service jobber restart

se_policy/.installed : ${SE_FILES}
	-[ -f /etc/sysconfig/selinux ] && ${MAKE} -C se_policy && semodule -i "$<" -v && restorecon -Rv /usr/local /etc/init.d
	touch "$@"

.PHONY : uninstall-bin
uninstall-bin :
	rm -f "${DESTDIR}/bin/${CLIENT}" "${DESTDIR}/sbin/${DAEMON}"
	-userdel "${CLIENT_USER}"

.PHONY : uninstall-centos
uninstall-centos :
	service jobber stop
	chkconfig jobber off
	chkconfig --del jobber
	rm -f /etc/init.d/jobber
	-[ -f /etc/sysconfig/selinux ] && semodule -r jobber -v
	rm -f se_policy/.installed

