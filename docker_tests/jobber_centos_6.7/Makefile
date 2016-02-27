IMAGE_NAME = jobber/centos6.7

.PHONY : test
test : image
	docker run "${IMAGE_NAME}"

.PHONY : image
image : Dockerfile src.tar
	docker build -t "${IMAGE_NAME}" .

.PHONY : src.tar
src.tar :
	cd ../.. && git archive --output "$@" HEAD
	mv "../../$@" ./

.PHONY : clean
clean :
	rm -f src.tar
