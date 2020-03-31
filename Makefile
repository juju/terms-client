# Copyright 2016 Canonical Ltd.
# Licensed under the GPLv3, see LICENCE file for details.
#
PROJECT := github.com/juju/terms-client

ifndef GOBIN
GOBIN := $(shell mkdir -p $(GOPATH)/bin; realpath $(GOPATH))/bin
else
REAL_GOBIN := $(shell mkdir -p $(GOBIN); realpath $(GOBIN))
GOBIN := $(REAL_GOBIN)
endif

ifeq ($(TERMS_CLIENT_SKIP_DEP),true)
dep:
	@echo "skipping dep"
else
$(GOPATH)/bin/dep:
	go get -u github.com/golang/dep/cmd/dep

# populate vendor/ from Gopkg.lock without updating it first (lock file is the single source of truth for machine).
dep: $(GOPATH)/bin/dep
	$(GOPATH)/bin/dep ensure -vendor-only $(verbose)
endif

# update Gopkg.lock (if needed), but do not update `vendor/`.
rebuild-dependencies:
	dep ensure -v -no-vendor $(dep-update)

build: dep
	GOBIN=$(GOBIN) go build -a $(PROJECT)/...

install: deps
	GOBIN=$(GOBIN) go install -v $(PROJECT)/...

check: build
	GOBIN=$(GOBIN) go test -test.timeout=1200s $(PROJECT)/...

land: check race

race: build
	GOBIN=$(GOBIN) go test -test.timeout=1200s -race $(PROJECT)/...

.PHONY: build check install clean race land 
