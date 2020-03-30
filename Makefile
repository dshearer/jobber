# GNU-standard vars (cf. http://www.gnu.org/prep/standards/html_node/Makefile-Conventions.html)
SHELL = /bin/sh
prefix = /usr/local
exec_prefix = ${prefix}
bindir = ${exec_prefix}/bin
libexecdir = ${exec_prefix}/libexec
sysconfdir = /etc
srcdir = .
INSTALL = install
INSTALL_PROGRAM = ${INSTALL}

JOBBER_VERSION := $(shell cat ${srcdir}/version)
SRC_TARBALL = jobber-${JOBBER_VERSION}.tgz
SRC_TARBALL_DIR = jobber-${JOBBER_VERSION}

OUTPUT_DIR = bin

GO = go
GO_VERSION = 1.11
# NOTE: '-mod=vendor' prevents go from downloading dependencies
GO_BUILD := ${GO} build -mod=vendor -ldflags "-X github.com/dshearer/jobber/common.jobberVersion=${JOBBER_VERSION}"
GO_VET = ${GO} vet -mod=vendor
GO_TEST = ${GO} test -mod=vendor
GO_GEN = ${GO_WITH_TOOLS} generate -mod=vendor
GO_CLEAN = ${GO} clean

PACKAGES = \
	github.com/dshearer/jobber/common \
	github.com/dshearer/jobber/ipc \
	github.com/dshearer/jobber/jobber \
	github.com/dshearer/jobber/jobbermaster \
	github.com/dshearer/jobber/jobberrunner \
	github.com/dshearer/jobber/jobfile

include mk/def-sources.mk

.PHONY : default
default : build

include mk/buildtools.mk # defines 'GO_WITH_TOOLS' and 'GOYACC'

.PHONY : build
build : ${OUTPUT_DIR}/jobber ${OUTPUT_DIR}/jobbermaster \
	${OUTPUT_DIR}/jobberrunner

.PHONY : check
check : ${TEST_SOURCES} jobfile/parse_time_spec.go
	@go version
	${GO_VET} ${PACKAGES}
	${GO_TEST} ${PACKAGES}

install : \
	${DESTDIR}${libexecdir}/jobbermaster \
	${DESTDIR}${libexecdir}/jobberrunner \
	${DESTDIR}${bindir}/jobber \
	${DESTDIR}${sysconfdir}/jobber.conf

${DESTDIR}${libexecdir}/% : ${OUTPUT_DIR}/%
	@echo INSTALL "$@"
	@mkdir -p "${dir $@}"
	@${INSTALL_PROGRAM} "$<" "$@"

${DESTDIR}${bindir}/% : ${OUTPUT_DIR}/%
	@echo INSTALL "$@"
	@mkdir -p "${dir $@}"
	@${INSTALL_PROGRAM} "$<" "$@"

${DESTDIR}${sysconfdir}/jobber.conf : ${OUTPUT_DIR}/jobbermaster
	@echo INSTALL "$@"
	@mkdir -p "${dir $@}"
	@"$<" defprefs > "$@"

.PHONY : uninstall
uninstall :
	-rm "${DESTDIR}${libexecdir}/jobbermaster"
	-rm "${DESTDIR}${libexecdir}/jobberrunner"
	-rm "${DESTDIR}${bindir}/jobber"
	-rm "${DESTDIR}${sysconfdir}/jobber.conf"

.PHONY : dist
dist :
	mkdir -p "${DESTDIR}dist-tmp"
	"${srcdir}/buildtools/srcsync" ${ALL_SOURCES} \
		"${DESTDIR}dist-tmp/${SRC_TARBALL_DIR}"
	tar -C "${DESTDIR}dist-tmp" -czf "${DESTDIR}${SRC_TARBALL}" \
		"${SRC_TARBALL_DIR}"
	rm -rf "${DESTDIR}dist-tmp"

${OUTPUT_DIR}/% : ${MAIN_SOURCES} jobfile/parse_time_spec.go
	@${srcdir}/buildtools/versionge "$$(go version | egrep -o '[[:digit:].]+' | head -n 1)" "${GO_VERSION}"
	@echo BUILD $*
	@${GO_BUILD} -o "$@" "github.com/dshearer/jobber/$*"

jobfile/parse_time_spec.go : ${GOYACC} ${JOBFILE_SOURCES}
	@echo GEN SRC
	@${GO_GEN} -mod=vendor github.com/dshearer/jobber/jobfile

.PHONY : clean
clean : clean-buildtools
	@echo CLEAN
	@-${GO_CLEAN} -i ${PACKAGES}
	@rm -rf "${DESTDIR}${SRC_TARBALL}.tgz" jobfile/parse_time_spec.go \
		jobfile/y.output "${OUTPUT_DIR}"
