PROJECT_NAME := "fog-computing-assignment"
PKG := "github.com/numyalai/$(PROJECT_NAME)"
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v /ext/ | grep -v _test.go)

all: clean build start

build: clean router client

router client: $(GO_FILES)
	@go build -o $@ -v $(PKG)/cmd/$@

start: router client
	./start.sh

clean:
	rm -f router client