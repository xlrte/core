

.PHONY: build-all-platforms
build-all-platforms: 
	docker buildx build --platform linux/amd64,linux/arm64 .