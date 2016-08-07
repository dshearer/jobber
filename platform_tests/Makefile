PLATFORMS = centos_6.7 alpine_3.3 debian_8.5

.PHONY : test
test :
	for P in ${PLATFORMS}; do make -C "$${P}"; done

.PHONY : clean
clean :
	for P in ${PLATFORMS}; do make -C "$${P}" clean; done