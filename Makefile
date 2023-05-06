.PHONY: release

release:
	goreleaser release --clean --snapshot
