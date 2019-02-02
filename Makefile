.PHONY : main
main :
	@echo "build - generate pages"
	@echo "relinfo - download release info from GitHub"
	@echo "serve - run local webserver"

.PHONY : build
build :
	php build.php

.PHONY : relinfo
relinfo :
	python3 get-release-info.py > phplib/latest-release.json

.PHONY : serve
serve :
	cd .. && serve .
