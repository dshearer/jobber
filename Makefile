# GNU-standard vars (cf. http://www.gnu.org/prep/standards/html_node/Makefile-Conventions.html)
SHELL = /bin/sh

# Paths may be loaded from mk/config.mk, whcih is made by configure
-include mk/config.mk
prefix ?= /usr/local
bindir ?= ${prefix}/bin
libexecdir ?= ${prefix}/libexec
sysconfdir ?= /etc
localstatedir ?= ${prefix}/var

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
GO_BUILD_BASE_FLAGS := -mod=vendor
COMPILETIME_VARS := \
	-X 'github.com/dshearer/jobber/common.jobberVersion=${JOBBER_VERSION}' \
	-X 'github.com/dshearer/jobber/common.etcDirPath=${sysconfdir}'

GO_BUILD := ${GO} build ${GO_BUILD_BASE_FLAGS} -ldflags "${COMPILETIME_VARS}"
GO_VET = ${GO} vet -mod=vendor
GO_TEST = ${GO} test -mod=vendor
GO_GEN = ${GO_WITH_TOOLS} generate -mod=vendor
GO_CLEAN = ${GO} clean -mod=vendor

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

################################################################################
# BUILD
################################################################################

.PHONY : build
build : check ${OUTPUT_DIR}/jobber ${OUTPUT_DIR}/jobbermaster \
		${OUTPUT_DIR}/jobberrunner ${OUTPUT_DIR}/jobber.conf
	@echo
	@echo "Built with these paths:"
	@echo "localstatedir: ${localstatedir}"
	@echo "libexecdir: ${libexecdir}"
	@echo "sysconfdir: ${sysconfdir}"

.PHONY : check
check : ${TEST_SOURCES} jobfile/parse_time_spec.go
	@go version
	@echo GO TEST
	@${GO_TEST} ${PACKAGES}

${OUTPUT_DIR}/% : ${MAIN_SOURCES} jobfile/parse_time_spec.go
	@$(call checkGoVersion)
	@echo GO VET
	@${GO_VET} ${PACKAGES}
	@echo BUILD $*
	@${GO_BUILD} -o "$@" "github.com/dshearer/jobber/$*"

${OUTPUT_DIR}/jobber.conf : ${OUTPUT_DIR}/jobbermaster
	@echo BUILD $@
	@${OUTPUT_DIR}/jobbermaster defprefs --var "${localstatedir}" --libexec "${libexecdir}" > "$@"

jobfile/parse_time_spec.go : ${GOYACC} ${JOBFILE_SOURCES}
	@echo GEN SRC
	@${GO_GEN} -mod=vendor github.com/dshearer/jobber/jobfile

################################################################################
# INSTALL
################################################################################

install : \
	${DESTDIR}${libexecdir}/jobbermaster \
	${DESTDIR}${libexecdir}/jobberrunner \
	${DESTDIR}${bindir}/jobber \
	${DESTDIR}${sysconfdir}/jobber.conf

${DESTDIR}${libexecdir}/% : ${OUTPUT_DIR}/%
	@echo INSTALL $@
	@mkdir -p "${dir $@}"
	@${INSTALL_PROGRAM} "$<" "$@"

${DESTDIR}${bindir}/% : ${OUTPUT_DIR}/%
	@echo INSTALL $@
	@mkdir -p "${dir $@}"
	@${INSTALL_PROGRAM} "$<" "$@"

${DESTDIR}${sysconfdir}/jobber.conf : ${OUTPUT_DIR}/jobber.conf
	@echo INSTALL $@
	@mkdir -p "${dir $@}"
	@cp "$<" "$@"

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

################################################################################
# CLEAN
################################################################################

.PHONY : clean
clean : clean-buildtools
	@echo CLEAN
	@-${GO_CLEAN} -i ${PACKAGES}
	@rm -rf "${DESTDIR}${SRC_TARBALL}.tgz" jobfile/parse_time_spec.go \
		jobfile/y.output "${OUTPUT_DIR}"

################################################################################
# MISC
################################################################################

define checkGoVersion
${srcdir}/buildtools/versionge "$$(go version | egrep -o '[[:digit:].]+' | head -n 1)" "${GO_VERSION}"
endef