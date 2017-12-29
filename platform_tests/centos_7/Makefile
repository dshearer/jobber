SRC_ROOT = ../..
IMAGE_NAME = jobber/centos7_test_rpm
TEST_SCRIPTS = $(wildcard ../test_scripts/*)

.PHONY : test
test : .made_rpms test_scripts.tar
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null

test_scripts.tar : ${TEST_SCRIPTS}
	tar -C .. -cf $@ test_scripts

.made_rpms :
	make -C "${SRC_ROOT}/packaging/centos_7" pkg-docker "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -f *.rpm test_scripts.tar docker_output .made_rpms
	make -C "${SRC_ROOT}/packaging/centos_7" clean
