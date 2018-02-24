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

GO_WKSPC ?= ${abspath ../../../..}
TEST_TMPDIR = ${PWD}
SRC_TARBALL = jobber-$(shell cat ${srcdir}/version).tgz
SRC_TARBALL_DIR = jobber-$(shell cat ${srcdir}/version)

GO = GOPATH=${GO_WKSPC} go

GO_VERSION = 1.8

LDFLAGS = -ldflags "-X github.com/dshearer/jobber/common.jobberVersion=`cat version`"

include mk/def-sources.mk

.PHONY : default
default : all

include mk/buildtools.mk

.PHONY : all
all : ${GO_WKSPC}/bin/jobber ${GO_WKSPC}/bin/jobbermaster \
	${GO_WKSPC}/bin/jobberrunner

.PHONY : check
check : ${TEST_SOURCES} jobfile/parse_time_spec.go
	@go version
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

install : \
	${DESTDIR}${libexecdir}/jobbermaster \
	${DESTDIR}${libexecdir}/jobberrunner \
	${DESTDIR}${bindir}/jobber \
	${DESTDIR}${sysconfdir}/jobber.conf

${DESTDIR}${libexecdir}/% : ${GO_WKSPC}/bin/%
	@echo INSTALL "$@"
	@mkdir -p "${dir $@}"
	@${INSTALL_PROGRAM} "$<" "$@"

${DESTDIR}${bindir}/% : ${GO_WKSPC}/bin/%
	@echo INSTALL "$@"
	@mkdir -p "${dir $@}"
	@${INSTALL_PROGRAM} "$<" "$@"

${DESTDIR}${sysconfdir}/jobber.conf : ${GO_WKSPC}/bin/jobbermaster
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

${GO_WKSPC}/bin/% : ${MAIN_SOURCES} jobfile/parse_time_spec.go
	@${srcdir}/buildtools/versionge "$$(go version | egrep --only-matching '[[:digit:].]+' | head -n 1)" "${GO_VERSION}"
	@echo BUILD $*
	@${GO} install ${LDFLAGS} "github.com/dshearer/jobber/$*"

jobfile/parse_time_spec.go : ${GOYACC} ${JOBFILE_SOURCES}
	@echo GEN SRC
	@${GO_WITH_TOOLS} generate github.com/dshearer/jobber/jobfile

.PHONY : clean
clean : clean-buildtools
	@echo CLEAN
	@-${GO} clean -i github.com/dshearer/jobber/common
	@-${GO} clean -i github.com/dshearer/jobber/jobfile
	@-${GO} clean -i github.com/dshearer/jobber/jobber
	@-${GO} clean -i github.com/dshearer/jobber/jobbermaster
	@-${GO} clean -i github.com/dshearer/jobber/jobberrunner
	@rm -f "${DESTDIR}${SRC_TARBALL}.tgz" jobfile/parse_time_spec.go \
		jobfile/y.output
