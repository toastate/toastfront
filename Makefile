# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
# +++ Check bash installed

DEPTEST=$(shell command -v bash 2> /dev/null)
ifeq ($(DEPTEST),)
$(error "Install bash to make it work")
endif


# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++


BUILDDIR                 := $(CURDIR)/build


# It's necessary to set this because some environments don't link sh -> bash.
SHELL                             = /usr/bin/env bash

GOPATH                            = $(shell go env GOPATH)
ARCH                              = $(shell uname -p)

GIT_COMMIT                        = $(shell git rev-parse HEAD)
GIT_SHA                           = $(shell git rev-parse --short HEAD)
GIT_TAG                           = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY                         = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")


# --------------------------------------------------------------------------------
# --------------------------------------------------------------------------------
# --------------------------------------------------------------------------------

.PHONY: all
all: build

.PHONY: build
build: $(BUILDDIR)
	@cd $(CURDIR)/cmd/toastfront && go build && mv -f toastfront $(BUILDDIR)

.PHONY: test
test: build
	@cd $(CURDIR)/example && $(BUILDDIR)/toastfront build

$(BUILDDIR):
	@mkdir -p $(BUILDDIR)