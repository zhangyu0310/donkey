GO_LDFLAGS ?=

GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.BuildTS=$(shell date '+%Y%m%d-%H%M')"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.CompileInfo=$(shell uname -p)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.GitHash=$(shell git rev-parse HEAD)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.GitBranch=$(shell git symbolic-ref --short -q HEAD)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.GitCommitDate=$(shell git log -1 --pretty=format:%ci)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.GOVersion=$(shell go version)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.MajorVersion=$(MAJOR_VERSION)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.MinorVersion=$(MINOR_VERSION)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.PatchVersion=$(PATCH_VERSION)"
GO_LDFLAGS += -X "$(ROOT_PACKAGE)/pkg/version.SuffixVersion=$(SUFFIX_VERSION)"

PROTOC ?= $(shell which protoc)

.PHONY: protoc
protoc: protoc-gen-go
	@(PATH=${PATH}:${ROOT_DIR}/hack/tools/bin protoc --proto_path=${ROOT_DIR}/pkg/proto --go_out=${ROOT_DIR}/pkg/proto --go_opt=paths=source_relative --go-grpc_out=${ROOT_DIR}/pkg/proto --go-grpc_opt=paths=source_relative ${ROOT_DIR}/pkg/proto/*.proto)

.PHONY: local.prebuild
local.prebuild:
	@(sleep 0)

.PHONY: local.postbuild
local.postbuild:
	@(sleep 0)
