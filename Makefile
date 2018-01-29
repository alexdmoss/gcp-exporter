export VERSION ?= $(shell git describe 2>/dev/null | sed -e 's/^v//g' || echo "dev")
export REVISION ?= $(shell git rev-parse --short HEAD || echo "unknown")
export BRANCH ?= $(shell git show-ref | grep "$(REVISION)" | grep -v HEAD | awk '{print $$2}' | sed 's|refs/remotes/origin/||' | sed 's|refs/heads/||' | sort | head -n 1)
export BUILT ?= $(shell date +%Y-%m-%dT%H:%M:%S%:z)
export CGO_ENABLED ?= 0

export CI_REGISTRY_IMAGE ?= registry.gitlab.com/gitlab-org/ci-cd/gcp-exporter
export CI_IMAGE ?= $(CI_REGISTRY_IMAGE)/ci:1.9-0

export TESTFLAGS ?= -cover

PKG = gitlab.com/gitlab-org/ci-cd/gcp-exporter
VERSION_PKG = $(PKG)/version

BUILD_DIR := $(CURDIR)

ORIGINAL_GOPATH = $(shell echo $$GOPATH)
LOCAL_GOPATH := $(CURDIR)/.gopath
GOPATH_SETUP := $(LOCAL_GOPATH)/.ok
GOPATH_BIN := $(LOCAL_GOPATH)/bin
PKG_BUILD_DIR := $(LOCAL_GOPATH)/src/$(PKG)

export GOPATH = $(LOCAL_GOPATH)
export PATH := $(GOPATH_BIN):$(PATH)

# Packages in vendor/ are included in ./...
# https://github.com/golang/go/issues/11659
export OUR_PACKAGES ?= $(subst _$(BUILD_DIR),$(PKG),$(shell go list ./... | grep -v -e '/vendor/' -e './cover/'))

GO_FILES ?= $(shell find . -name '*.go' | grep -v './.gopath/')

GO_LDFLAGS := -X $(VERSION_PKG).VERSION=$(VERSION) \
              -X $(VERSION_PKG).REVISION=$(REVISION) \
              -X $(VERSION_PKG).BRANCH=$(BRANCH) \
              -X $(VERSION_PKG).BUILT=$(BUILT) \
              -s -w

# Development Tools
DEP = $(GOPATH_BIN)/dep
MOCKERY = $(GOPATH_BIN)/mockery

MOCKERY_FLAGS = -note="This comment works around https://github.com/vektra/mockery/issues/155"

.PHONY: version
version:
	@echo Current version: $(VERSION)
	@echo Current revision: $(REVISION)
	@echo Current branch: $(BRANCH)

.PHONY: goinfo
goinfo: $(DEP)
	#
	# ORIGINAL_GOPATH
	@echo $(ORIGINAL_GOPATH)
	# LOCAL_GOPATH
	@echo $(LOCAL_GOPATH)
	# GOPATH_BIN
	@echo $(GOPATH_BIN)
	# PKG_BUILD_DIR
	@echo $(PKG_BUILD_DIR)
	# GOPATH
	@echo $(GOPATH)
	# PATH
	@echo $(PATH)
	# dep version
	@$(DEP) version

.PHONY: deps
deps: $(DEP)
	# Installing required dependencies
	@cd $(PKG_BUILD_DIR) && dep ensure

.PHONY: codequality
codequality:
	# Checking code quality
	@./scripts/codequality analyze -f json --dev | tee codeclimate.json

.PHONY: fmt
fmt: $(GOPATH_SETUP)
	# Fixing project code formatting...
	@go fmt $(OUR_PACKAGES) | awk '{if (NF > 0) {if (NR == 1) print "Please run go fmt for:"; print "- "$$1}} END {if (NF > 0) {if (NR > 0) exit 1}}'

.PHONY: test
test: deps
	# Running unit tests (with: CGO=$(CGO_ENABLED) flags=$(TESTFLAGS))
	@./scripts/go_test_with_coverage_report

.PHONY: race_test
race_test: CGO_ENABLED=1
race_test: TESTFLAGS=-cover -race
race_test: test

.PHONY: check_race_conditions
check_race_conditions:
	# Checking if there is an increase of detected race conditions
	@./scripts/check_race_conditions

.PHONY: compile
compile: deps
	# Compile binary
	@go build -ldflags "$(GO_LDFLAGS)" $(PKG)
	# Test run
	@./gcp-exporter --version

.PHONY: release_image
release_image:
	# Release Docker image
	@./scripts/release_image

.PHONY: release_ci_image
release_ci_image:
	# Release CI Docker image
	@./scripts/release_ci_image

# We rely on user GOPATH 'cause mockery seems not to be able to find dependencies in vendor directory
.PHONY: mocks
mocks: $(MOCKERY)
	@find . -type f -name 'mock_*' -delete
	@GOPATH=$(ORIGINAL_GOPATH) mockery $(MOCKERY_FLAGS) -dir=./client -all -inpkg
	@GOPATH=$(ORIGINAL_GOPATH) mockery $(MOCKERY_FLAGS) -dir=./collectors -all -inpkg
	@GOPATH=$(ORIGINAL_GOPATH) mockery $(MOCKERY_FLAGS) -dir=./services -all -inpkg
	@GOPATH=$(ORIGINAL_GOPATH) mockery $(MOCKERY_FLAGS) -dir=./tests -all -inpkg

#
# local GOPATH setup
#

$(GOPATH_SETUP): $(PKG_BUILD_DIR)
	# Creating GOPATH_BIN directory ($(GOPATH_BIN))
	@mkdir -p $(GOPATH_BIN)
	@touch $@

$(PKG_BUILD_DIR):
	# Creating PKG_BUILD directory ($(PKG_BUILD_DIR))
	@mkdir -p $(@D)
	# Linking sources in PKG_BUILD directory
	@ln -s ../../../../.. $@

#
# development tools setup
#

$(DEP): $(GOPATH_SETUP)
    # installing github.com/golang/dep/cmd/dep ($(DEP))
	@CGO_ENABLED=1 go get github.com/golang/dep/cmd/dep

$(MOCKERY): $(GOPATH_SETUP)
	# installing github.com/vektra/mockery/.../ ($(MOCKERY))
	@CGO_ENABLED=1 go get github.com/vektra/mockery/.../

.PHONY: clean
clean:
	# Removing LOCAL_GOPATH ($(LOCAL_GOPATH))
	@-$(RM) -rf $(LOCAL_GOPATH)
