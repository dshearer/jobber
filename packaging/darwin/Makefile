include ../head.mk

# required by tail.mk:
PKGFILE = nop
PKGFILE_DEPS = nop
PKGFILE_VM_PATH = nop
PACKAGING_SUBDIR = darwin
SRC_TARBALL_DIR = jobber
INSTALL_PKG_CMD = nop
UNINSTALL_PKG_CMD = nop

.PHONY : test-local
test-local : ${WORK_DIR}/${SRC_TARBALL}
	./make-and-install.sh ${WORK_DIR}/${SRC_TARBALL} $${PWD}

include ../tail.mk
