# GNU-standard vars (cf. http://www.gnu.org/prep/standards/html_node/Makefile-Conventions.html)
SHELL = /bin/sh
prefix = /usr/local
exec_prefix = ${prefix}
bindir = ${exec_prefix}/bin
libexecdir = ${exec_prefix}/libexec
srcdir = .
INSTALL = install
INSTALL_PROGRAM = ${INSTALL}

GO_WKSPC ?= ${abspath ../../../..}
LIB = jobber.a
TEST_TMPDIR = ${PWD}
DIST_PKG_NAME = jobber-$(shell cat ${srcdir}/version)

GO = GO15VENDOREXPERIMENT=1 GOPATH=${GO_WKSPC} go
GODEP = GO15VENDOREXPERIMENT=1 GOPATH=${GO_WKSPC} godep

# read lists of source files
include common/sources.mk \
		jobber/sources.mk \
		jobbermaster/sources.mk \
		jobberrunner/sources.mk \
		jobfile/sources.mk \
		packaging/sources.mk
FINAL_LIB_SOURCES := \
	$(COMMON_SOURCES:%=common/%) \
	$(JOBFILE_SOURCES:%=jobfile/%)
FINAL_LIB_TEST_SOURCES := \
	$(COMMON_TEST_SOURCES:%=common/%) \
	$(JOBFILE_TEST_SOURCES:%=jobfile/%)
FINAL_CLIENT_SOURCES := $(CLIENT_SOURCES:%=jobber/%)
FINAL_CLIENT_TEST_SOURCES := $(CLIENT_TEST_SOURCES:%=jobber/%)
FINAL_MASTER_SOURCES := $(MASTER_SOURCES:%=jobbermaster/%)
FINAL_MASTER_TEST_SOURCES := $(MASTER_TEST_SOURCES:%=jobbermaster/%)
FINAL_RUNNER_SOURCES := $(RUNNER_SOURCES:%=jobberrunner/%)
FINAL_RUNNER_TEST_SOURCES := $(RUNNER_TEST_SOURCES:%=jobberrunner/%)
FINAL_PACKAGING_SOURCES := $(PACKAGING_SOURCES:%=packaging/%)

MAIN_SOURCES := \
	${FINAL_LIB_SOURCES} \
	${FINAL_CLIENT_SOURCES} \
	${FINAL_MASTER_SOURCES} \
	${FINAL_RUNNER_SOURCES}
	
TEST_SOURCES := \
	${FINAL_LIB_TEST_SOURCES} \
	${FINAL_CLIENT_TEST_SOURCES} \
	${FINAL_MASTER_TEST_SOURCES} \
	${FINAL_RUNNER_TEST_SOURCES}

GO_SOURCES := \
	${MAIN_SOURCES} \
	${TEST_SOURCES}
	
OTHER_SOURCES := \
	Makefile \
	common/sources.mk \
	jobber/sources.mk \
	jobbermaster/sources.mk \
	jobberrunner/sources.mk \
	jobfile/sources.mk \
	packaging/sources.mk \
	buildtools \
	README.md \
	LICENSE \
	version \
	Godeps \
	vendor \
	${FINAL_PACKAGING_SOURCES}

ALL_SOURCES := \
	${GO_SOURCES} \
	${OTHER_SOURCES}

LDFLAGS = -ldflags "-X github.com/dshearer/jobber/common.jobberVersion=`cat version`"

SE_FILES = se_policy/jobber.fc \
           se_policy/jobber.if \
           se_policy/jobber.te \
           ${wildcard se_policy/include/**} \
           se_policy/Makefile \
           se_policy/policygentool

.PHONY : all
all : lib ${GO_WKSPC}/bin/jobber ${GO_WKSPC}/bin/jobbermaster ${GO_WKSPC}/bin/jobberrunner

.PHONY : check
check : ${TEST_SOURCES}
	${GO} vet \
		github.com/dshearer/jobber/common \
		github.com/dshearer/jobber/jobber \
		github.com/dshearer/jobber/jobbermaster \
		github.com/dshearer/jobber/jobberrunner \
		github.com/dshearer/jobber/jobfile
	TMPDIR="${TEST_TMPDIR}" ${GO} test \
		github.com/dshearer/jobber/common \
		github.com/dshearer/jobber/jobber \
		github.com/dshearer/jobber/jobbermaster \
		github.com/dshearer/jobber/jobberrunner \
		github.com/dshearer/jobber/jobfile

.PHONY : installcheck
installcheck :
	./test_installation

.PHONY : installdirs
installdirs :
	"${srcdir}/buildtools/mkinstalldirs" "${DESTDIR}${bindir}" "${DESTDIR}${libexecdir}"

.PHONY : install
install : installdirs all
	# install files
	"${INSTALL_PROGRAM}" "${GO_WKSPC}/bin/jobbermaster" "${DESTDIR}${libexecdir}"
	"${INSTALL_PROGRAM}" "${GO_WKSPC}/bin/jobberrunner" "${DESTDIR}${libexecdir}"
	"${INSTALL_PROGRAM}" "${GO_WKSPC}/bin/jobber" "${DESTDIR}${bindir}"

.PHONY : uninstall
uninstall :
	-rm "${DESTDIR}${libexecdir}/jobbermaster"
	-rm "${DESTDIR}${libexecdir}/jobberrunner"
	-rm "${DESTDIR}${bindir}/jobber"

dist : ${ALL_SOURCES}
	mkdir -p "${DESTDIR}dist-tmp"
	"${srcdir}/buildtools/srcsync" ${ALL_SOURCES} \
		"${DESTDIR}dist-tmp/${DIST_PKG_NAME}"
	tar -C "${DESTDIR}dist-tmp" -czf "${DESTDIR}${DIST_PKG_NAME}.tgz" \
		"${DIST_PKG_NAME}"
	rm -rf "${DESTDIR}dist-tmp"

.PHONY : clean
clean :
	-${GO} clean -i github.com/dshearer/jobber/common
	-${GO} clean -i github.com/dshearer/jobber/jobfile
	-${GO} clean -i github.com/dshearer/jobber/jobber
	-${GO} clean -i github.com/dshearer/jobber/jobbermaster
	-${GO} clean -i github.com/dshearer/jobber/jobberrunner
	rm -f "${DESTDIR}${DIST_PKG_NAME}.tgz"

.PHONY : lib
lib : ${FINAL_LIB_SOURCES}
	@go version
	${GO} install ${LDFLAGS} "github.com/dshearer/jobber/common"
	${GO} install ${LDFLAGS} "github.com/dshearer/jobber/jobfile"

${GO_WKSPC}/bin/jobber : ${FINAL_CLIENT_SOURCES} lib
	${GO} install ${LDFLAGS} github.com/dshearer/jobber/jobber

${GO_WKSPC}/bin/jobbermaster : ${FINAL_MASTER_SOURCES} lib
	${GO} install ${LDFLAGS} github.com/dshearer/jobber/jobbermaster

${GO_WKSPC}/bin/jobberrunner : ${FINAL_RUNNER_SOURCES} lib
	${GO} install ${LDFLAGS} github.com/dshearer/jobber/jobberrunner

.PHONY : get-deps
get-deps :
	${GODEP} save ./...

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

.PHONY : uninstall-centos
uninstall-centos :
	service jobber stop
	chkconfig jobber off
	chkconfig --del jobber
	rm -f /etc/init.d/jobber
	-[ -f /etc/sysconfig/selinux ] && semodule -r jobber -v
	rm -f se_policy/.installed

