PLATFORMS_DOCKER = centos_6 alpine_3.3 debian_8 ubuntu_16.04 centos_7
PLATFORMS_VM = centos_6 debian_8 ubuntu_17 centos_7

.PHONY : main
main :
	@echo "Choose test-docker or test-vm or clean"

.PHONY : test-docker
test-docker :
	for P in ${PLATFORMS_DOCKER}; do make -C "$${P}" test-docker || exit 1; done

.PHONY : test-vm
test-vm :
	for P in ${PLATFORMS_VM}; do make -C "$${P}" test-vm || exit 1; done

.PHONY : clean
clean :
	for P in ${PLATFORMS_DOCKER}; do make -C "$${P}" clean; done
	for P in ${PLATFORMS_VM}; do make -C "$${P}" clean; done