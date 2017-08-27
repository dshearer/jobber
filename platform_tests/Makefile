PLATFORMS = centos_6.7 alpine_3.3 debian_8 ubuntu_17 centos_7

.PHONY : test
test :
	for P in ${PLATFORMS}; do make -C "$${P}" || exit 1; done

.PHONY : clean
clean :
	for P in ${PLATFORMS}; do make -C "$${P}" clean; done