

.PHONY: build-all-platforms
build-all-platforms: 
	docker buildx build --platform linux/amd64,linux/arm64 .

.PHONY: pre-commit
pre-commit: lint test
	go test ./...

.PHONY: test
test: plugins
	go test ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: plugins
plugins:
	mkdir -p pkg/api/testdata/plugins
	go build -buildmode=plugin -o pkg/api/testdata/plugins/gcp.so plugins/gcp/gcp.go
