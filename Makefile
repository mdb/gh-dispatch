SOURCE=./...
GOFMT_FILES?=$$(find . -type f -name '*.go')
VERSION?=0.1.5

default: build

tools:
	go install github.com/goreleaser/goreleaser@v1.11.4
.PHONY: tools

build: tools
	goreleaser release \
		--snapshot \
		--skip-publish \
		--rm-dist
.PHONY: build

test: vet fmtcheck
	go test -v -coverprofile=coverage.out -count=1 $(SOURCE)
.PHONY: test

acc-test:
	go test -v --tags=acceptance $(SOURCE)
.PHONY: acceptance-test

vet:
	go vet $(SOURCE)
.PHONY: vet

fmt:
	gofmt -w $(GOFMT_FILES)
.PHONY: fmt

fmtcheck:
	test -z $(shell go fmt $(SOURCE))
.PHONY: fmtcheck

check-tag:
	./scripts/ensure-unique-version.sh "$(VERSION)"
.PHONY: check-tag

tag: check-tag
	echo "creating git tag $(VERSION)"
	git tag $(VERSION)
	git push origin $(VERSION)
.PHONY: tag

release: tools
	goreleaser release \
		--rm-dist
.PHONY: release

# TODO: dynamically set architecture, which is currently hard-coded to amd64
# And why is a dist/gh-dispatch_linux_amd64_v1/gh-dispatch (note the _v1) produced in GHA?
install:
	mkdir -p ~/.local/share/gh/extensions/gh-dispatch
	cp dist/gh-dispatch_$(shell echo $(shell uname) | tr '[:upper:]' '[:lower:]')_amd64*/gh-dispatch ~/.local/share/gh/extensions/gh-dispatch/
.PHONY: install

demo:
	vhs < demo.tape
.PHONY: demo
