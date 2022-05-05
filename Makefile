

.PHONY: lint-golang
lint-golang:
	$(QUIET) scripts/check-go-fmt.sh
	$(QUIET) $(GO_VET)  ./...
	$(QUIET) golangci-lint run


