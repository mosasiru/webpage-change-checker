VERBOSE_FLAG = $(if $(VERBOSE),-v)

VERSION = $$(git describe --tags --always --dirty) ($$(git name-rev --name-only HEAD))

BUILD_FLAGS = -ldflags "\
	      -X \"main.Version=$(VERSION)\" \
	      "
build: deps
	go build $(VERBOSE_FLAG) $(BUILD_FLAGS)

deps:
	go get -d $(VERBOSE_FLAG)

install: deps
	go install $(VERBOSE_FLAG) $(BUILD_FLAGS)

bump-minor:
	git diff --quiet && git diff --cached --quiet
	new_version=$$(gobump minor -w -r -v) && \
	git commit -a -m "bump version to $$new_version" && \
	git tag v$$new_version

.PHONY: build deps install
