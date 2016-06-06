PLATFORMS = centos_6.7 alpine_3.3

.PHONY : pkg-docker
pkg-docker :
	for P in ${PLATFORMS}; do make -C "$${P}" pkg-docker; done

.PHONY : clean
clean :
	for P in ${PLATFORMS}; do make -C "$${P}" clean; done