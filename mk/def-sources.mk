# read lists of source files
include buildtools/sources.mk \
		common/sources.mk \
		jobber/sources.mk \
		jobbermaster/sources.mk \
		jobberrunner/sources.mk \
		jobfile/sources.mk \
		packaging/sources.mk

MAIN_SOURCES := \
	${COMMON_SOURCES} \
	${JOBFILE_SOURCES} \
	${CLIENT_SOURCES} \
	${MASTER_SOURCES} \
	${RUNNER_SOURCES}

TEST_SOURCES := \
	${COMMON_TEST_SOURCES} \
	${JOBFILE_TEST_SOURCES} \
	${CLIENT_TEST_SOURCES} \
	${MASTER_TEST_SOURCES} \
	${RUNNER_TEST_SOURCES}

GO_SOURCES := \
	${MAIN_SOURCES} \
	${TEST_SOURCES}

OTHER_SOURCES := \
	.circleci \
	Godeps \
	LICENSE \
	Makefile \
	mk \
	platform_tests \
	README.md \
	smoke_tests/Dockerfile \
	smoke_tests/Makefile \
	.travis.yml \
	vendor \
	version \
	${PACKAGING_SOURCES} \
	${BUILDTOOLS_SOURCES}

ALL_SOURCES := \
	${GO_SOURCES} \
	${OTHER_SOURCES}
