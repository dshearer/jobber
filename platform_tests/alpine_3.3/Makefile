SRC_ROOT = ../..
IMAGE_NAME = jobber/alpine3.3_test_pkg

.PHONY : test
test : .made_pkgs
	cp "${SRC_ROOT}/test_installation" ./
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "Jobber installation is good." docker_output

.made_pkgs :
	make -C "${SRC_ROOT}/packaging/alpine_3.3" pkg-docker "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -f *.apk test_installation docker_output .made_pkgs
	make -C "${SRC_ROOT}/packaging/alpine_3.3" clean
