SRC_ROOT = ../..
IMAGE_NAME = jobber/alpine3.6_test_pkg
TEST_SCRIPTS := $(wildcard ../test_scripts/*)

.PHONY : test
test : .made_pkgs test_scripts.tar
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "platform_tests: pass" docker_output > /dev/null

test_scripts.tar : ${TEST_SCRIPTS}
	tar -C .. -cf $@ test_scripts

.made_pkgs :
	make -C "${SRC_ROOT}/packaging/alpine_3.6" pkg-docker "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -f *.apk test_scripts.tar docker_output .made_pkgs
	make -C "${SRC_ROOT}/packaging/alpine_3.6" clean
