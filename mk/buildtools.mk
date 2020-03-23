_BUILDTOOLS_GEN = ${CURDIR}/buildtools/gen
_TOOLS_TAR = tools-v0.3.4.tar.gz
_TOOLS_DIR = tools-gopls-v0.3.4

GOYACC = ${_BUILDTOOLS_GEN}/bin/goyacc
GO_WITH_TOOLS = PATH="${_BUILDTOOLS_GEN}/bin:$${PATH}" ${GO}

${GOYACC} : buildtools/gotools/${_TOOLS_TAR}
	@echo MAKE GOYACC
	@cd buildtools/gotools/ && tar -xzf "${_TOOLS_TAR}"
	@cd "buildtools/gotools/${_TOOLS_DIR}/cmd/goyacc" && ${GO_BUILD} -o "${GOYACC}"

.PHONY : clean-buildtools
clean-buildtools :
	@rm -rf "${_BUILDTOOLS_GEN}" "buildtools/gotools/${_TOOLS_DIR}"
