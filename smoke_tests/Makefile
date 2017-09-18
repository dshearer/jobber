DOCKER_IMAGE_NAME = jobber/smoke_test
PLATFORM = centos_7
SRC_ROOT := $(abspath ..)
VERSION := $(shell cat ${SRC_ROOT}/version)
SRC_TARFILE = jobber.tgz

DOCKER_RUN = docker run ${DOCKER_IMAGE_NAME}

.PHONY : test-docker
test-docker : platform_tests.tar jobber.tgz
	docker build -t "${DOCKER_IMAGE_NAME}" .
	(docker run "${DOCKER_IMAGE_NAME}" ||:) > testlog.txt
	mkdir -p test_report
	docker cp "`docker ps -alq`:/log.html" test_report/
	docker cp "`docker ps -alq`:/report.html" test_report/
	
	@cat testlog.txt
	@egrep '.* critical tests,.* 0 failed[[:space:]]*$$' testlog.txt\
		>/dev/null

platform_tests.tar : $(wildcard ${SRC_ROOT}/platform_tests/**)
	tar -C "${SRC_ROOT}" -cf "$@" platform_tests

jobber.tgz :
	make -C "${SRC_ROOT}" dist "DESTDIR=$(abspath .)/"
	mv jobber-${VERSION}.tgz "$@"

.PHONY : clean
clean :
	rm -rf platform_tests.tar jobber.tgz test_report testlog.txt