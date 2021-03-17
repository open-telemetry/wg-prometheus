# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Use local registry for kind
IMG_REG = localhost:5000
IMG_VER = 0.1
IMG ?= $(IMG_REG)/clb:$(IMG_VER)

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

PROJECT_ROOT := $(shell pwd)
BIN := $(PROJECT_ROOT)/bin
CONTROLLER_GEN = $(BIN)/controller-gen
KUSTOMIZE = $(BIN)/kustomize

.PHONY: tools proto

manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run goimports against code.
	$(BIN)/goimports -d -l -w ./cmd ./configmanager ./controllers ./loadbalancer

vet: ## Run go vet against code.
	go vet ./...

tools: ## Install tools, version controlled under hack/tools
	(export GOBIN=$(PROJECT_ROOT)/bin && cd hack/tools && ./install.sh)

proto: ## Generates protobuf
	(export PATH=$(BIN):$(PATH) && cd proto && protoc --proto_path=src --go_out=./generated/clb --go_opt module=github.com/aws-observability/collector-load-balancer/proto/generated/clb clb.proto)
	(export PATH=$(BIN):$(PATH) && cd proto && protoc -I ./src --proto_path=src --go_out=./generated/clb/configmanager \
		--go_opt=module=github.com/aws-observability/collector-load-balancer/proto/generated/clb/configmanager \
		--go-grpc_out=./generated/clb/configmanager --go-grpc_opt=module=github.com/aws-observability/collector-load-balancer/proto/generated/clb/configmanager \
		configmanager/configmanager.proto)
	(export PATH=$(BIN):$(PATH) && cd proto && protoc -I ./src --proto_path=src --go_out=./generated/clb/loadbalancer \
		--go_opt module=github.com/aws-observability/collector-load-balancer/proto/generated/clb/loadbalancer \
		--go-grpc_out=./generated/clb/loadbalancer --go-grpc_opt=module=github.com/aws-observability/collector-load-balancer/proto/generated/clb/loadbalancer \
		loadbalancer/loadbalancer.proto)
	(export PATH=$(BIN):$(PATH) && cd proto && protoc -I ./src --proto_path=src --grpc-gateway_out=./generated/clb/configmanager\
		--grpc-gateway_opt=logtostderr=true \
		--grpc-gateway_opt=module=github.com/aws-observability/collector-load-balancer/proto/generated/clb/configmanager \
		--grpc-gateway_opt=grpc_api_configuration=src/configmanager/configmanager_http.yaml \
		configmanager/configmanager.proto)

##@ Build

GOBUILD = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build

build-clbotelcol:
	cd cmd/clbotelcol && $(GOBUILD) .

build-clb:
	$(GOBUILD) -o cmd/clb/clb main.go

build: generate fmt vet build-clbotelcol build-clb ## Build manager, collector,  binary.
# TODO: sample app

run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

clb-docker: build-clb ## Build and push controller image
	cd cmd/clb && docker build -t $(IMG) .
	docker push $(IMG)

clbotelcol-docker: build-clbotelcol
	cd cmd/clbotelcol &&  docker build -t $(IMG_REG)/clbotelcol:$(IMG_VER) .
	docker push $(IMG_REG)/clbotelcol:$(IMG_VER)

##@ Deployment

install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests clb-docker ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -
