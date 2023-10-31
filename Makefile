ROOT_PACKAGE=$(shell grep -e '^module\s\+\(.\+\)$$' go.mod|sed -E 's/module[[:blank:]]*//')

MAJOR_VERSION:=1
MINOR_VERSION:=0
PATCH_VERSION:=1
SUFFIX_VERSION:=$(if $(shell git status --porcelain 2>/dev/null), dirty, )

IMAGE_REPOSITORY:=zhangyu0310

all: fix build image
default: build

## libs
include hack/makelib/common.mk
include hack/makelib/golang.mk
include hack/makelib/image.mk
include hack/makelib/tools.mk
## targets
include hack/local/local.mk
include hack/makelib/ui.mk
