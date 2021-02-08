BUILDDIR = build
BUILDFLAGS =

APPS = nsq_to_elasticsearch
all: $(APPS)

$(BUILDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BUILDFLAGS} -o $@ ./

$(APPS): %: $(BUILDDIR)/%

clean:
	rm -rf $(BUILDDIR)

test:
	go test -v -race -cover -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: clean all test lint
.PHONY: $(APPS)

lint:
	golangci-lint run --tests=false ./...
