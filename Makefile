# Copyright 2021 The TiPrometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le s390x


GOLANGCI_LINT_OPTS ?= --timeout 2m

include Makefile.common


DOCKER_IMAGE_NAME       ?= tiprometheus


.PHONY: test
test: common-test

.PHONY: docker
docker: common-docker

.PHONY: build
build: common-all
	@echo ">> building binaries"
	$(GO) build -o tiprometheus cmd/tiprometheus/app.go
