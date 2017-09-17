DOCKER_IMAGE_NAME = jobber/smoke_test
PLATFORM = centos_7
SRC_ROOT := $(abspath ..)
VERSION := $(shell cat ${SRC_ROOT}/version)
SRC_TARFILE = jobber.tgz
ROBOT_TESTS_DIR = ${SRC_ROOT}/platform_tests/robot_tests
ROBOT_TESTS = $(wildcard ${ROBOT_TESTS_DIR}/*)

DOCKER_RUN = docker run ${DOCKER_IMAGE_NAME}

.PHONY : test-docker
test-docker : robot_tests.tar jobber.tgz
	docker build -t "${DOCKER_IMAGE_NAME}" .
	docker run "${DOCKER_IMAGE_NAME}"
	mkdir -p test_report
	docker cp "`docker ps -alq`:/log.html" test_report/
	docker cp "`docker ps -alq`:/report.html" test_report/

robot_tests.tar : ${ROBOT_TESTS}
	tar -C ${ROBOT_TESTS_DIR}/.. -cf "$@" robot_tests

jobber.tgz :
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/"
	mv jobber-${VERSION}.tgz "$@"

.PHONY : clean
clean :
	rm -rf robot_tests.tar jobber.tgz test_report