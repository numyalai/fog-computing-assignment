PROJECT_NAME := "fog-computing-assignment"
PKG := "github.com/numyalai/$(PROJECT_NAME)"
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v /ext/ | grep -v _test.go)

all: clean build start

build: router client cpu_watcher ram_watcher

router client cpu_watcher ram_watcher: $(GO_FILES)
	@go build -o ./services/$@ -v $(PKG)/cmd/$@

start: router client cpu_watcher ram_watcher
	./start.sh

clean:
	rm -rf ./services