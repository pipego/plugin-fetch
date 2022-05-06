# Config

VERSION=$(version)


# Build

.PHONY: FORCE

build: go-build
.PHONY: build

clean: go-clean
.PHONY: clean

lint: go-lint
.PHONY: lint


# Non-PHONY targets (real files)

go-build: FORCE
	./script/build.sh $(VERSION)

go-clean: FORCE
	./script/clean.sh

go-lint: FORCE
	./script/lint.sh
