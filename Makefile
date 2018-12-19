SRCS = $(shell git ls-files '*.go')
.PHONY: test

all:
	go build
clean:
	go clean
test:
	go test
bench:
	go test -bench . -benchmem	
fmt:
	gofmt -w $(SRCS)
