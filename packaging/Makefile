PLATFORMS_DOCKER = centos_6 alpine_3.3 debian_8 centos_7
PLATFORMS_VM = centos_6 debian_8 centos_7

.PHONY : main
main :
	@echo "Choose pkg-docker or pkg-vm or clean"

.PHONY : pkg-docker
pkg-docker :
	for P in ${PLATFORMS_DOCKER}; do make -C "$${P}" pkg-docker; done

.PHONY : pkg-vm
pkg-vm :
	for P in ${PLATFORMS_VM}; do make -C "$${P}" pkg-vm; done

.PHONY : clean
clean :
	for P in ${PLATFORMS_DOCKER}; do make -C "$${P}" clean; done
	for P in ${PLATFORMS_VM}; do make -C "$${P}" clean; done