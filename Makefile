GO111MODULE := on
DOCKER_TAG := $(or ${GIT_TAG_NAME}, latest)

all: eventrouter

.PHONY: eventrouter
eventrouter:
	go build -tags netgo -o bin/eventrouter *.go
	strip bin/eventrouter

.PHONY: dockerimages
dockerimages:
	docker build -t mwennrich/eventrouter:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
	docker push mwennrich/eventrouter:${DOCKER_TAG}

.PHONY: clean
clean:
	rm -f bin/*
