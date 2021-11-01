
git_tag=$(shell git describe --tags $(git rev-list --tags --max-count=1))
date=$(shell date)

.PHONY: pre-commit
pre-commit: lint test
	go test ./...

.PHONY: test
test:
	go clean -testcache
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build-mac-universal
build-mac-universal:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'github.com/xlrte/core/pkg/cmd.Version=$(git_tag)' -X 'github.com/xlrte/core/pkg/cmd.BuildDate=\"$(date)\"'" -o xlrte-mac-x86 cmd/cli/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X 'github.com/xlrte/core/pkg/cmd.Version=$(git_tag)' -X 'github.com/xlrte/core/pkg/cmd.BuildDate=\"$(date)\"'" -o xlrte-mac-arm64 cmd/cli/main.go
	lipo -create -output xlrte-mac-universal xlrte-mac-x86 xlrte-mac-arm64
	rm xlrte-mac-x86 xlrte-mac-arm64

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X 'github.com/xlrte/core/pkg/cmd.Version=$(git_tag)' -X 'github.com/xlrte/core/pkg/cmd.BuildDate=\"$(date)\"'" -o xlrte-linux-x86 cmd/cli/main.go

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X 'github.com/xlrte/core/pkg/cmd.Version=$(git_tag)' -X 'github.com/xlrte/core/pkg/cmd.BuildDate=\"$(date)\"'" -o xlrte-windows-x86 cmd/cli/main.go

.PHONY: build
build: build-mac-universal build-linux build-windows

# .PHONY: plugins
# plugins:
# 	mkdir -p pkg/api/testdata/pluginsz
# 	go build -buildmode=plugin -trimpath -o pkg/api/testdata/plugins/gcp.so plugins/gcp/gcp.go
