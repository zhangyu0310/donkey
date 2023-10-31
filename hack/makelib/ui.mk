.PHONY: fmt
fmt:
	@$(MAKE) go.source.gofmt
	@$(MAKE) go.source.gci

.PHONY: fix
fix: fmt
	@$(MAKE) go.mod.tidy
	@$(MAKE) go.lint.fix
	@$(MAKE) go.source.licenses

.PHONY: check
check: fmt go.mod.tidy
	@git --no-pager diff --exit-code || (echo "Please add changed files!" && false)
	@$(MAKE) go.lint

# test: Run unit test.
.PHONY: test
test:
	@$(MAKE) go.test

# build: Build source code for host platform.
.PHONY: build
build:
	@$(MAKE) prebuild
	@$(MAKE) go.build
	@$(MAKE) postbuild

.PHONY: build.%
build.%:
	@$(MAKE) go.build.$*

.PHONY: build-debug
build-debug:
	@$(MAKE) go.build DEBUG=true

.PHONY: build-debug.%
build-debug.%:
	@$(MAKE) go.build.$* DEBUG=true


# image: Build docker images
.PHONY: image
image: prebuild image.build

.PHONY: image-debug
image-debug:
	@$(MAKE) image.build DEBUG=true


.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)

#$(info ROOT_DIR=$(ROOT_DIR))
.PHONY: dbg
dbg:
	@echo '##variables##'
	@echo '# ROOT_DIR=$(ROOT_DIR)'
	@echo '# VERSION=$(VERSION)'
	@echo '# ROOT_PACKAGE=$(ROOT_PACKAGE)'
	@echo '# TOOLS_DIR="$(TOOLS_DIR)"'
	@echo '# IMAGE_REPOSITORY=$(IMAGE_REPOSITORY)'
	@echo '# IMAGES=$(IMAGES)'
	@echo '# BINS=$(BINS)'
	@echo '# GO_LDFLAGS=$(GO_LDFLAGS)'
	@echo

.PHONY: prebuild
prebuild:
	@$(MAKE) local.prebuild

.PHONY: postbuild
postbuild:
	@$(MAKE) local.postbuild

