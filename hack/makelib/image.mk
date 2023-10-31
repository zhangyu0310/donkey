IMAGE_REPOSITORY ?= #local

IMAGES_DIR ?= $(wildcard $(ROOT_DIR)/build/*)

IMAGES ?= $(foreach image,$(IMAGES_DIR),$(notdir $(image)))
TAGEXT ?=
TAG ?=

IMAGE_REPO_PREFIX := $(IMAGE_REPOSITORY)

ifneq (${IMAGE_REPOSITORY},)
ifneq ($(word $(words $(IMAGE_REPOSITORY)),$(IMAGE_REPOSITORY)), /)
	IMAGE_REPO_PREFIX := $(IMAGE_REPO_PREFIX)/
endif
endif

ifeq (${IMAGES},)
	$(error Could not determine IMAGES, set ROOT_DIR or run in source dir)
endif

.PHONY: image.build
image.build: $(addprefix image.build., $(IMAGES))

.PHONY: image.push
image.push: $(addprefix image.push., $(IMAGES))

.PHONY: image.build.%
image.build.%: go.build.%.linux_amd64
	$(eval _NAME:=$*)
	$(eval _BUILD_TMP_DIR:=$(TMP_DIR)/$(_NAME)-docker)
	@echo "===========> Building docker image ($(_NAME):$(VERSION))"
	$(eval _TAG:=$(VERSION)$(if $(filter $(DEBUG),true),-dev)$(if $(TAGEXT),-$(TAGEXT),))
	$(eval _HOOK_DIRS:=$(ROOT_DIR)/build/$(_NAME)/)
	@rm -rf $(_BUILD_TMP_DIR)
	@mkdir -p $(_BUILD_TMP_DIR)/bin
	@cp ${OUTPUT_DIR}/linux/amd64/$(_NAME) $(_BUILD_TMP_DIR)/bin/ || true

	@for hookdir in $(_HOOK_DIRS); do \
	(if [[ -x $${hookdir}/pre-build.sh ]]; then \
	  cd $${hookdir}; \
	  DST_DIR=$(_BUILD_TMP_DIR) \
	  ROOT_DIR=$(ROOT_DIR) \
		$${hookdir}/pre-build.sh; \
	fi) ;\
	done

#   @docker build --pull
	@docker build \
		-t $(IMAGE_REPO_PREFIX)$(_NAME):$(_TAG) \
		-f $(ROOT_DIR)/build/$(_NAME)/Dockerfile $(_BUILD_TMP_DIR)

ifneq ($(strip $(TAG)),)
	@docker tag $(IMAGE_REPO_PREFIX)$(_NAME):$(_TAG) $(IMAGE_REPO_PREFIX)$(_NAME):$(TAG) && \
	echo "tagged $(IMAGE_REPO_PREFIX)$(_NAME):$(TAG)"
endif

	@for hookdir in $(_HOOK_DIRS); do \
	(if [[ -x $${hookdir}/post-build.sh ]]; then \
	  cd $${hookdir}; \
	  ROOT_DIR=$(ROOT_DIR) \
	  DST_DIR=$(_BUILD_TMP_DIR) \
	  OUTPUT_DIR=$(OUTPUT_DIR)/ \
	  IMAGE_REPOSITORY=$(IMAGE_REPOSITORY) \
	  IMAGE_NAME=$(_NAME) \
	  IMAGE_CANONICAL_TAG=$(_TAG) \
	  IMAGE_TAGS="$(_TAG) $(TAG)" \
	  $${hookdir}/post-build.sh; \
	fi) ;\
	done

.PHONY: image.push.%
image.push.%: image.build.%
	$(eval _NAME:=$*)
	$(eval _TAG:=$(VERSION)$(if $(filter $(DEBUG),true),-dev,$(if $(TAGEXT),-$(TAGEXT),)))
	@echo "===========> Pushing docker image ($(IMAGE_REPO_PREFIX)$(_NAME):$(_TAG))"
	docker push $(IMAGE_REPO_PREFIX)$(_NAME):$(_TAG)
ifneq ($(strip $(TAG)),)
	docker push $(IMAGE_REPO_PREFIX)$(_NAME):$(TAG)
endif
