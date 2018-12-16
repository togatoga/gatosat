SRCS = $(shell git ls-files '*.go')

all:
	go build
clean:
	go clean
bench:
	go test -bench . -benchmem	
fmt:
	gofmt -w $(SRCS)
