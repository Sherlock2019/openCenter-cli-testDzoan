.PHONY: clean lint rke terraform kubectl

BIN := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/.bin
TERRAFORM_VERSION := 1.12.2
KUBECTL_VERSION := 1.28.0
HELM_VERSION := 3.18.6

export PATH := $(BIN):$(PATH)
export TF_CLI_CONFIG_FILE=config.tfrc

export ANSIBLE_INVENTORY = {{- if .OpenCenter.GitOps.GitDir }}{{ .OpenCenter.GitOps.GitDir }}/inventory/inventory.yaml{{- else }}/tmp/inventory/inventory.yaml{{- end }}

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_S),Linux)
	OS = linux
	ifeq ($(UNAME_M),x86_64)
		ARCH = amd64
	endif
	ifeq ($(UNAME_M),aarch64)
		ARCH = arm64
	endif
endif
ifeq ($(UNAME_S),Darwin)
	OS = darwin
	ifeq ($(UNAME_M),x86_64)
		ARCH = amd64
	endif
	ifeq ($(UNAME_M),arm64)
		ARCH = arm64
	endif
endif

clean:
	rm cluster.rkestate kube_config_cluster.yml terraform.tfstate*

rke:

terraform:
	@if ! terraform --version | head -n 1 | grep $(TERRAFORM_VERSION); then \
		mkdir -p $(BIN); \
		curl -L https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/terraform_$(TERRAFORM_VERSION)_$(OS)_$(ARCH).zip > $(BIN)/terraform.zip; \
		unzip $(BIN)/terraform.zip -d $(BIN); \
		rm $(BIN)/terraform.zip; \
	fi;

kubectl:
	@if ! kubectl version --client --output=yaml 2>/dev/null | grep -q "gitVersion: v$(KUBECTL_VERSION)"; then \
		mkdir -p $(BIN); \
		curl -L "https://dl.k8s.io/release/v$(KUBECTL_VERSION)/bin/$(OS)/$(ARCH)/kubectl" -o $(BIN)/kubectl; \
		chmod +x $(BIN)/kubectl; \
	fi;

helm:
	@if ! helm version --template="{{.Version}}" 2>/dev/null | grep -q "v$(HELM_VERSION)"; then \
		mkdir -p $(BIN); \
		curl -L "https://get.helm.sh/helm-v$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz" | tar xz -C $(BIN) --strip-components=1 $(OS)-$(ARCH)/helm; \
	fi;
