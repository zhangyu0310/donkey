DEBUG ?= false

PLATFORM ?= $(shell go env GOOS)_$(shell go env GOARCH)
COMMANDS ?= $(wildcard ${ROOT_DIR}/cmd/*)
BINS ?= $(foreach cmd,${COMMANDS},$(notdir ${cmd}))

FN_SET_PKG_VERSION = \
-X $(1).gitVersion=$(VERSION) \
-X $(1).gitCommit=$(GIT_COMMIT) \
-X $(1).gitTreeState=$(GIT_TREE_STATE)
#-X $(1).buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

GO_MIN_VERSION := 1.14
GO_GCFLAGS ?=
GO_LDFLAGS += $(foreach pkg,$(VERSION_PACKAGES),$(call FN_SET_PKG_VERSION,$(pkg)))
GO_TAGS ?=
ifeq ($(RELEASE),true)
ifeq ($(DEBUG),true)
$(error "RELEASE and DEBUG flag shouldn't be set simultaneously")
endif
GO_LDFLAGS +=-w -s
endif

GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
GOARCH := $(word 2, $(subst _, ,$(PLATFORM)))
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
CGO_ENABLED ?= 0

ifeq ($(GOOS),windows)
GO_OUT_EXT := .exe
endif

.PHONY: go.build
go.build: $(addprefix go.build., $(addsuffix .$(PLATFORM), $(BINS)))

.PHONY: go.build.verify
go.build.verify:
ifeq ($(COMMANDS),)
	$(error Could not determine COMMANDS, no $$(ROOT_DIR) set or empty in $(ROOT_DIR)/cmd/)
endif
ifneq ($(shell printf '%s\ngo%s\n' $$(go version | grep -oE 'go[0-9.]+') $(GO_MIN_VERSION) | sort -rcV; echo $$?), 0)
	$(error go version must not less than $(GO_MIN_VERSION))
endif

#FN_GCFLAGS = $(if $(filter $(DEBUG),true)
.PHONY: go.build.%
go.build.%: go.build.verify
	$(eval COMMAND := $(word 1,$(subst ., ,$*)))
	$(eval _PLATFORM := $(word 2,$(subst ., ,$*)))
	$(eval _PLATFORM := $(if $(_PLATFORM),$(_PLATFORM),$(PLATFORM)))
	$(eval OS := $(word 1,$(subst _, ,$(_PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(_PLATFORM))))
	@echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)"
	@mkdir -p $(OUTPUT_DIR)/$(OS)/$(ARCH)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(OS) GOARCH=$(ARCH) \
		go build $(if $(filter $(DEBUG),true),-gcflags all="-N -l",-trimpath) \
		$(if $(strip $(GO_GCFLAGS)),-gcflags '$(GO_GCFLAGS)',) \
		$(if $(strip $(GO_TAGS)),-tags '$(GO_TAGS)',) \
		-ldflags '$(GO_LDFLAGS)' \
		-o $(OUTPUT_DIR)/$(OS)/$(ARCH)/$(COMMAND)$(GO_OUT_EXT) \
		./cmd/$(COMMAND)

.PHONY: go.lint
go.lint: golangci-lint
	@echo "===========> Run golangci-lint"
	@hack/tools/bin/golangci-lint run

.PHONY: go.lint.fix
go.lint.fix: golangci-lint
	@echo "===========> Run golangci-lint with autofix"
	@hack/tools/bin/golangci-lint run --fix

.PHONY: go.mod.tidy
go.mod.tidy:
	@echo "===========> Run go mod tidy"
	@go mod tidy

.PHONY: go.test
go.test:
	@echo "===========> Run $@"
	go test -timeout=10m -short -v -gcflags=-l -cover ./...

.PHONY: go.source.gofmt
go.source.gofmt:
	@echo "===========> Run gofmt"
	@gofmt -w .

.PHONY: go.source.gci
go.source.gci: gci
	@echo "===========> Run gci (format imports)"
	@hack/tools/bin/gci write .

.PHONY: go.source.licenses
go.source.licenses: go-licenses
	@echo "===========> Run go-licenses (check licenses)"
	@hack/tools/bin/go-licenses check $(ROOT_PACKAGE)/pkg/... --allowed_licenses ",Apache-1.0,Apache-1.1,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC,MIT,MPL-2.0,LGPL-3.0,PHP,Python,ZLIB,LIBPNG" --logtostderr=false --stderrthreshold FATAL
