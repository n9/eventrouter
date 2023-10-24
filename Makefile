GO111MODULE := on

all: eventrouter

.PHONY: eventrouter
eventrouter:
	go build -tags netgo -o bin/eventrouter *.go
	strip bin/eventrouter

.PHONY: clean
clean:
	rm -f bin/*
