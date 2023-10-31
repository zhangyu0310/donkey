.PHONY: gci goimports go-licenses

gci: hack/tools/bin/gci
golangci-lint: hack/tools/bin/golangci-lint
go-licenses: hack/tools/bin/go-licenses
protoc-gen-go: hack/tools/bin/protoc-gen-go

GOBUILD=GOWORK=off go build -mod=mod

hack/tools/bin/gci: hack/tools/go.mod
	@cd hack/tools && $(GOBUILD) -o ./bin/gci github.com/daixiang0/gci

hack/tools/bin/golangci-lint: hack/tools/go.mod
	@cd hack/tools && $(GOBUILD) -o ./bin/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

hack/tools/bin/go-licenses: hack/tools/go.mod
	@cd hack/tools && $(GOBUILD) -o ./bin/go-licenses github.com/google/go-licenses

hack/tools/bin/protoc-gen-go: hack/tools/go.mod
	@cd hack/tools && $(GOBUILD) -o ./bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go
