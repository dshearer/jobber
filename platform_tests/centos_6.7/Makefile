SRC_ROOT = ../..
IMAGE_NAME = jobber/centos6.7_test_rpm

.PHONY : test
test : .made_rpms
	cp "${SRC_ROOT}/test_installation" ./
	docker build -t "${IMAGE_NAME}" .
	docker run "${IMAGE_NAME}" | tee docker_output
	@grep "Jobber installation is good." docker_output

.made_rpms :
	make -C "${SRC_ROOT}/packaging/centos_6.7" pkg-docker "DESTDIR=$(abspath .)/"
	touch "$@"

.PHONY : clean
clean :
	rm -f *.rpm test_installation docker_output .made_rpms
