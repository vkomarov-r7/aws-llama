run:
	go run . serve

.PHONY: release-check
release-check:
	goreleaser release --snapshot --clean

.PHONY: release
release:
	goreleaser release --clean
