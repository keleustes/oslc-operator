
# Image URL to use all building/pushing image targets
COMPONENT        ?= openstacklcm-operator
VERSION_V3       ?= 3.11.0
DHUBREPO         ?= keleustes/${COMPONENT}-dev
DOCKER_NAMESPACE ?= keleustes
IMG_V2           ?= ${DHUBREPO}:v${VERSION_V2}
IMG_V3           ?= ${DHUBREPO}:v${VERSION_V3}

all: docker-build

setup:
# ifndef GOPATH
# 	$(error GOPATH not defined, please define GOPATH. Run "go help gopath" to learn more about GOPATH)
# endif
	echo $(GOPATH)

.PHONY: clean
clean:
	rm -fr vendor
	rm -fr cover.out
	rm -fr build/_output
	rm -fr config/crds

.PHONY: install-tools
install-tools:
	cd /tmp && GO111MODULE=on go get sigs.k8s.io/kind@v0.5.0
	cd /tmp && GO111MODULE=on go get github.com/instrumenta/kubeval@0.13.0

# clusterexist=$(shell kind get clusters | grep oslc  | wc -l)
# ifeq ($(clusterexist), 1)
#   testcluster=$(shell kind get kubeconfig-path --name="oslc")
#   SETKUBECONFIG=KUBECONFIG=$(testcluster)
# else
#   SETKUBECONFIG=
# endif

.PHONY: which-cluster
which-cluster:
	echo $(SETKUBECONFIG)

.PHONY: create-testcluster
create-testcluster:
	kind create cluster --name oslc

.PHONY: delete-testcluster
delete-testcluster:
	kind delete cluster --name oslc


# Run tests
unittest: setup fmt vet-v3
	echo "sudo systemctl stop kubelet"
	echo -e 'docker stop $$(docker ps -qa)'
	echo -e 'export PATH=$${PATH}:/usr/local/kubebuilder/bin'
	mkdir -p config/crds
	cp chart/templates/*v1alpha1* config/crds/
	GO111MODULE=on go test ./pkg/... ./cmd/... -coverprofile cover.out

# Run go fmt against code
fmt: setup
	GO111MODULE=on go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet-v3: fmt
	GO111MODULE=on go vet -composites=false -tags=v3 ./pkg/... ./cmd/...

# Generate code
generate: setup
        # git clone sigs.k8s.io/controller-tools
        # go install ./cmd/...
	GO111MODULE=on controller-gen crd paths=./pkg/apis/openstacklcm/... crd:trivialVersions=true output:crd:dir=./chart/templates/ output:none
	GO111MODULE=on controller-gen object paths=./pkg/apis/openstacklcm/... output:object:dir=./pkg/apis/openstacklcm/v1alpha1 output:none

# Build the docker image
docker-build: fmt docker-build-v3

docker-build-v3: vet-v3
	GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/_output/bin/openstacklcm-operator -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} -tags=v3 ./cmd/...
	docker build . -f build/Dockerfile -t ${IMG_V3}
	docker tag ${IMG_V3} ${DHUBREPO}:latest


# Push the docker image
docker-push: docker-push-v3

docker-push-v3:
	docker push ${IMG_V3}

# Run against the configured Kubernetes cluster in ~/.kube/config
install: install-v3

purge: setup
	helm delete --purge openstacklcm-operator

install-v3: docker-build-v3
	helm install --name openstacklcm-operator chart --set images.tags.operator=${IMG_V3}

# Deploy and purge procedure which do not rely on helm itself
install-kubectl: setup
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_controllerrevisions.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_deletephases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_installphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_operationalphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_oslcs.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_planningphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_rollbackphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_testphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_trafficdrainphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_trafficrolloutphases.yaml
	kubectl apply -f ./chart/templates/openstacklcm.airshipit.org_upgradephases.yaml
	kubectl apply -f ./chart/templates/role_binding.yaml
	kubectl apply -f ./chart/templates/roles.yaml
	kubectl apply -f ./chart/templates/service_account.yaml
	kubectl apply -f ./chart/templates/argo_openstacklcm_role.yaml
	kubectl create -f deploy/operator.yaml

purge-kubectl: setup
	kubectl delete -f deploy/operator.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_controllerrevisions.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_deletephases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_installphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_operationalphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_oslcs.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_planningphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_rollbackphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_testphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_trafficdrainphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_trafficrolloutphases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/openstacklcm.airshipit.org_upgradephases.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/role_binding.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/roles.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/service_account.yaml --ignore-not-found=true
	kubectl delete -f ./chart/templates/argo_openstacklcm_role.yaml --ignore-not-found=true

getcrds:
	kubectl get oslcs.openstacklcm.airshipit.org

	kubectl get installphases.openstacklcm.airshipit.org
	kubectl get rollbackphases.openstacklcm.airshipit.org
	kubectl get testphases.openstacklcm.airshipit.org
	kubectl get trafficdrainphases.openstacklcm.airshipit.org
	kubectl get trafficrolloutphases.openstacklcm.airshipit.org
	kubectl get upgradephases.openstacklcm.airshipit.org
	kubectl get deletephases.openstacklcm.airshipit.org
	kubectl get planningphases.openstacklcm.airshipit.org
	kubectl get operationalphases.openstacklcm.airshipit.org

	kubectl get workflows.argoproj.io

getcrddetails:
	kubectl get -o yaml oslcs.openstacklcm.airshipit.org

	kubectl get -o yaml installphases.openstacklcm.airshipit.org
	kubectl get -o yaml rollbackphases.openstacklcm.airshipit.org
	kubectl get -o yaml testphases.openstacklcm.airshipit.org
	kubectl get -o yaml trafficdrainphases.openstacklcm.airshipit.org
	kubectl get -o yaml trafficrolloutphases.openstacklcm.airshipit.org
	kubectl get -o yaml upgradephases.openstacklcm.airshipit.org
	kubectl get -o yaml deletephases.openstacklcm.airshipit.org
	kubectl get -o yaml planningphases.openstacklcm.airshipit.org
	kubectl get -o yaml operationalphases.openstacklcm.airshipit.org
