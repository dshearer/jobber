PLATFORMS = centos_6 alpine_3.6 debian_9 centos_7

.PHONY : pkg-docker
pkg-docker :
	for P in ${PLATFORMS}; do make -C "$${P}" pkg-docker; done

.PHONY : clean
clean :
	for P in ${PLATFORMS}; do make -C "$${P}" clean; done