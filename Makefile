run:
	go run . serve

.PHONY: release-check
release-check:
	goreleaser release --snapshot --clean

.PHONY: release
release:
	goreleaser release --clean

install: release-check
	cp dist/aws-llama_darwin_amd64_v1/aws-llama /usr/local/bin/aws-llama
