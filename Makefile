########################################################################
## Development

# Custom version suffix for use during development. The default, the
# empty string, ensures that CI, releases, etc. are not affected by
# this additional variable. During development on the other hand it
# can be used to provide a string allowing developers to distinguish
# different binaries built from the same base commit. A simple means
# for that would be a timestamp. For example, via
#
# VSUFFIX="-$(date +%Y-%m-%dT%H-%M-%S)" make ...
#
# yielding, for example
#
# % ep version
# Epinio Version: "v0.1.6-16-ge5ad0849-2021-11-18T10-00-27"

VSUFFIX ?= 
VERSION ?= $(shell git describe --tags)$(VSUFFIX)
CGO_ENABLED ?= 0
export LDFLAGS += -X github.com/epinio/epinio/internal/version.Version=$(VERSION)

build: build-amd64

build-win: build-windows

build-all: build-amd64 build-arm64 build-arm32 build-windows build-darwin build-darwin-m1

build-all-small:
	@$(MAKE) LDFLAGS+="-s -w" build-all

build-linux-arm: build-arm32
build-arm32:
	GOARCH="arm" GOOS="linux" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-linux-arm32

build-linux-arm64: build-arm64
build-arm64:
	GOARCH="arm64" GOOS="linux" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-linux-arm64

build-linux-amd64: build-amd64
build-amd64:
	GOARCH="amd64" GOOS="linux" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-linux-amd64

build-windows-amd64: build-windows
build-windows:
	GOARCH="amd64" GOOS="windows" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-windows-amd64.exe

build-darwin-amd64: build-darwin
build-darwin:
	GOARCH="amd64" GOOS="darwin" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-darwin-amd64

build-darwin-arm64: build-darwin-m1
build-darwin-m1:
	GOARCH="arm64" GOOS="darwin" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-darwin-arm64

build-linux-s390x: build-s390x
build-s390x:
	GOARCH="s390x" GOOS="linux" CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_ARGS) -ldflags '$(LDFLAGS)' -o dist/epinio-linux-s390x

build-images: build-linux-amd64
	@./scripts/build-images.sh

compress:
	upx --brute -1 ./dist/epinio-linux-arm32
	upx --brute -1 ./dist/epinio-linux-arm64
	upx --brute -1 ./dist/epinio-linux-amd64
	upx --brute -1 ./dist/epinio-windows-amd64.exe
	upx --brute -1 ./dist/epinio-darwin-amd64
	upx --brute -1 ./dist/epinio-darwin-arm64

test:
	ginkgo --nodes ${GINKGO_NODES} -r -p -race --fail-on-pending helpers internal

tag:
	@git describe --tags --abbrev=0

########################################################################
# Acceptance tests

FLAKE_ATTEMPTS ?= 2
GINKGO_NODES ?= 2
GINKGO_SLOW_TRESHOLD ?= 200
REGEX ?= ""

acceptance-cluster-delete:
	k3d cluster delete epinio-acceptance
	@if test -f /usr/local/bin/rke2-uninstall.sh; then sudo sh /usr/local/bin/rke2-uninstall.sh; fi

acceptance-cluster-delete-kind:
	kind delete cluster --name epinio-acceptance

acceptance-cluster-setup:
	@./scripts/acceptance-cluster-setup.sh

acceptance-cluster-setup-kind:
	@./scripts/acceptance-cluster-setup-kind.sh

test-acceptance: showfocus
	ginkgo --nodes ${GINKGO_NODES} --slow-spec-threshold ${GINKGO_SLOW_TRESHOLD}s --randomize-all --flake-attempts=${FLAKE_ATTEMPTS} --fail-on-pending acceptance/. acceptance/api/v1/. acceptance/apps/.

test-acceptance-api: showfocus
	ginkgo --nodes ${GINKGO_NODES} --slow-spec-threshold ${GINKGO_SLOW_TRESHOLD}s --randomize-all --flake-attempts=${FLAKE_ATTEMPTS} --fail-on-pending acceptance/api/v1/.

test-acceptance-apps: showfocus
	ginkgo --nodes ${GINKGO_NODES} --slow-spec-threshold ${GINKGO_SLOW_TRESHOLD}s --randomize-all --flake-attempts=${FLAKE_ATTEMPTS} --fail-on-pending acceptance/apps/.

test-acceptance-cli: showfocus
	ginkgo --nodes ${GINKGO_NODES} --slow-spec-threshold ${GINKGO_SLOW_TRESHOLD}s --randomize-all --flake-attempts=${FLAKE_ATTEMPTS} --fail-on-pending acceptance/.

test-acceptance-install: showfocus
	# TODO support for labels is coming in ginkgo v2
	ginkgo --nodes ${GINKGO_NODES} --focus "${REGEX}" --randomize-all --flake-attempts=${FLAKE_ATTEMPTS} acceptance/install/.

showfocus:
	@if test `cat acceptance/*.go acceptance/apps/*.go acceptance/api/v1/*.go | grep -c 'FIt\|FWhen\|FDescribe\|FContext'` -gt 0 ; then echo ; echo 'Focus:' ; grep 'FIt\|FWhen\|FDescribe\|FContext' acceptance/*.go acceptance/apps/*.go acceptance/api/v1/*.go ; echo ; fi

generate:
	go generate ./...

# Assumes that the `docs` checkout is a sibling of the `epinio` checkout
generate-cli-docs:
	@./scripts/cli-docs-generate.sh ../docs/src/references/cli

lint:
	go vet ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

check:
	golangci-lint run

patch-epinio-deployment:
	@./scripts/patch-epinio-deployment.sh

########################################################################
# Docs

getswagger:
	( [ -x "$$(command -v swagger)" ] || go install github.com/go-swagger/go-swagger/cmd/swagger@v0.28.0 )

swagger: getswagger
	swagger generate spec > docs/references/api/swagger.json
	sed -i 's/^{/{ "info": {"title": "Epinio", "version":"1"},/' docs/references/api/swagger.json
	swagger validate        docs/references/api/swagger.json

swagger-serve: getswagger
	swagger serve docs/references/api/swagger.json

########################################################################
# Support

tools-install:
	@./scripts/tools-install.sh

tools-versions:
	@./scripts/tools-versions.sh

########################################################################
# Kube dev environments

minikube-start:
	@./scripts/minikube-start.sh

minikube-delete:
	@./scripts/minikube-delete.sh

install-cert-manager:
	helm repo add cert-manager https://charts.jetstack.io
	helm repo update
	echo "Installing Cert Manager"
	helm upgrade --install cert-manager --create-namespace -n cert-manager \
		--set installCRDs=true \
		--set extraArgs[0]=--enable-certificate-owner-ref=true \
		cert-manager/cert-manager --version 1.7.1 \
		--wait

prepare_environment_k3d: build-linux-amd64 build-images
	@./scripts/prepare-environment-k3d.sh

unprepare_environment_k3d:
	kubectl delete --ignore-not-found=true secret regcred
	helm uninstall epinio -n epinio --wait || true
