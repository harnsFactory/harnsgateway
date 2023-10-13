DBG_MAKEFILE ?=
ifeq ($(DBG_MAKEFILE),1)
	$(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
	$(warning ***** $(shell date))
else
	# If we're not debugging the Makefile, don't echo recipes.
	MAKEFLAGS += -s
endif

# Old-skool build tools.
#
# Commonly used targets (see each target for more information):
#   all: Build code.
#   test: Run tests.
#   clean: Clean up.

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash

# Define variables so `make --warn-undefined-variables` works.
PRINT_HELP ?=

# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

# Constants used throughout.
.EXPORT_ALL_VARIABLES:
OUT_DIR ?= _output
BIN_DIR := $(OUT_DIR)/bin
GOPATH?=$(shell go env GOPATH)

# This controls the verbosity of the build.  Higher numbers mean more output.
KUBE_VERBOSE ?= 1

define ALL_HELP_INFO
# Build code.
#
# Args:
#   WHAT: Directory names to build.  If any of these directories has a 'main'
#     package, the build will produce executable files under $(OUT_DIR)/bin.
#     If not specified, "everything" will be built.
#   GOFLAGS: Extra flags to pass to 'go' when building.
#   GOLDFLAGS: Extra linking flags passed to 'go' when building.
#   GOGCFLAGS: Additional go compile flags passed to 'go' when building.
#
# Example:
#   make
#   make all
#   make all WHAT=cmd/harnsgateway GOFLAGS=-v
#   make all GOLDFLAGS=""
#     Note: Specify GOLDFLAGS as an empty string for building unstripped binaries, which allows
#           you to use code debugging tools like delve. When GOLDFLAGS is unspecified, it defaults
#           to "-s -w" which strips debug information. Other flags that can be used for GOLDFLAGS 
#           are documented at https://golang.org/cmd/link/
endef
.PHONY: all
ifeq ($(PRINT_HELP),y)
all:
	@echo "$$ALL_HELP_INFO"
else
all: verify-golang
	hack/make-rules/build.sh $(WHAT)
endif

define CHECK_TEST_HELP_INFO
# Build and run tests.
#
# Args:
#   WHAT: Directory names to test.  All *_test.go files under these
#     directories will be run.  If not specified, "everything" will be tested.
#   TESTS: Same as WHAT.
#   KUBE_COVER: Whether to run tests with code coverage. Set to 'y' to enable coverage collection.
#   GOFLAGS: Extra flags to pass to 'go' when building.
#   GOLDFLAGS: Extra linking flags to pass to 'go' when building.
#   GOGCFLAGS: Additional go compile flags passed to 'go' when building.
#
# Example:
#   make check
#   make test
#   make check WHAT=./pkg/harnsgateway GOFLAGS=-v
endef

.PHONY: verify-golang
verify-golang:
	hack/verify-golang.sh
