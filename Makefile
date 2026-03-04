export PATH := $(GO_BIN):$(PATH)

GO_BIN?=$(shell pwd)/.bin
GOCI_LINT_VERSION?=v2.10.1

lint-breaking::
	buf breaking --against 'https://github.com/DmitriiPro/go-grpc-buf-project.git#branch=main'

lint-proto::
	buf lint --config buf.yaml

format::
	"$(GO_BIN)/golangci-lint" run --fix -v ./...

generate-proto::
	buf generate --template buf.gen.yaml

install-tools::
	@echo "Installing tools..."
	@mkdir -p "$(GO_BIN)"
	curl -sSfl https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(GO_BIN)" $(GOCI_LINT_VERSION)
	

tidy::
	go mod tidy -v