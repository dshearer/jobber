PLATFORMS = centos_6 debian_8 centos_7 ubuntu_16

.PHONY : main
main :
	@echo "Choose pkg-vm or test-vm or clean or deepclean"

define RECIPES
.PHONY : pkg-vm-${platform}
pkg-vm-${platform} :
	mkdir -p "results/${platform}"
	$${MAKE} -C ${platform} pkg-vm "DESTDIR=$(abspath results/${platform})/"

test-vm-${platform} :
	mkdir -p "results/${platform}"
	$${MAKE} -C ${platform} test-vm "DESTDIR=$(abspath results/${platform})/"

clean-${platform} :
	$${MAKE} -C ${platform} clean

deepclean-${platform} :
	$${MAKE} -C ${platform} deepclean
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