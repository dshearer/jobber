PLATFORMS = centos_6 centos_7 alpine_3.6 debian_9 ubuntu_16.04

.PHONY : test
test :
	for P in ${PLATFORMS}; do make -C "$${P}" || exit 1; done

.PHONY : clean
clean :
	for P in ${PLATFORMS}; do make -C "$${P}" clean; done