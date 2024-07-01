PROJECT_NAME := "fog-computing-assignment"
PKG := "github.com/numyalai/$(PROJECT_NAME)"
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v /ext/ | grep -v _test.go)

all: clean build start

build: router client watcher

router client watcher: $(GO_FILES)
	@go build -o ./services/$@ -v $(PKG)/cmd/$@

start: router client watcher
	./start.sh

clean:
	rm -rf ./services