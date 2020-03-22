_BUILDTOOLS_GEN = ${CURDIR}/buildtools/gen

GOYACC = ${_BUILDTOOLS_GEN}/bin/goyacc
GO_WITH_TOOLS = PATH="${_BUILDTOOLS_GEN}/bin:$${PATH}" ${GO}

${GOYACC} : buildtools/gotools/tools-v0.3.4.tar.gz
	@echo MAKE GOYACC
	@cd buildtools/gotools/ && tar -xzf tools-v0.3.4.tar.gz
	@cd buildtools/gotools/tools-gopls-v0.3.4/cmd/goyacc && go build -mod=vendor -o "${GOYACC}"

.PHONY : clean-buildtools
clean-buildtools :
	@rm -rf "${_BUILDTOOLS_GEN}" "buildtools/gotools/${_TOOLS_DIR}"
