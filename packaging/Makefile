PLATFORMS = centos_6 centos_7 debian_9 ubuntu_16

.PHONY : main
main :
	@echo "Choose pkg-vm or test-vm or clean or deepclean"

define RECIPES
.PHONY : pkg-vm-${platform}
pkg-vm-${platform} :
	mkdir -p "results/${platform}"
	$${MAKE} -j 1 -C ${platform} pkg-vm \
		"DESTDIR=$(abspath results/${platform})/" \
	    2>&1 | tee "results/${platform}/pkg-vm.log"

test-vm-${platform} :
	mkdir -p "results/${platform}"
	$${MAKE} -j 1 -C ${platform} test-vm \
		"DESTDIR=$(abspath results/${platform})/" \
	    2>&1 | tee "results/${platform}/test-vm.log"

clean-${platform} :
	$${MAKE} -j 1 -C ${platform} clean

deepclean-${platform} :
	$${MAKE} -j 1 -C ${platform} deepclean
endef

$(foreach platform,${PLATFORMS}, $(eval ${RECIPES}))

.PHONY : pkg-vm
pkg-vm : $(patsubst %,pkg-vm-%,${PLATFORMS})

.PHONY : test-vm
test-vm : $(patsubst %,test-vm-%,${PLATFORMS})

.PHONY : clean
clean : $(patsubst %,clean-%,${PLATFORMS})
	rm -rf results

.PHONY : deepclean
deepclean : $(patsubst %,deepclean-%,${PLATFORMS})