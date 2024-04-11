HELM_NAME = e2e
HELM_FILE = config/local-config.yaml

HELM_REPO = starship
HELM_CHART = devnet
HELM_VERSION = v0.1.38

###############################################################################
###                              All commands                               ###
###############################################################################

.PHONY: setup
setup: setup-deps setup-helm setup-kind

.PHONY: stop
stop: stop-forward delete

.PHONY: clean
clean: stop clean-kind

.PHONY: test
test:
	go test -v ./...


###############################################################################
###                          Dependency check                               ###
###############################################################################

.PHONY: check
setup-deps:
	bash $(CURDIR)/scripts/dev-setup.sh

###############################################################################
###                              Helm Charts                                ###
###############################################################################

# run command with `$ make install FILE=configs/osmo-juno-cosmos.yaml`
..PHONY: setup-helm
setup-helm:
	helm repo add $(HELM_REPO) https://cosmology-tech.github.io/starship/
	helm repo update
	helm search repo $(HELM_REPO)/$(HELM_CHART) --version $(HELM_VERSION)

.PHONY: install
install:
	@echo "Installing the helm chart. This is going to take a while....."
	@echo "You can check the status with \"kubectl get pods\", run in another terminal please"
	helm install  $(HELM_NAME) $(HELM_REPO)/$(HELM_CHART) --version $(HELM_VERSION) -f $(HELM_FILE)
	@echo "Run \"kubectl get pods\" to check the status of the cluster"

.PHONY: upgrade
upgrade:
	helm upgrade --debug -f $(HELM_FILE) $(HELM_NAME) $(HELM_REPO)/$(HELM_CHART) --version $(HELM_VERSION)

.PHONY: debug
debug:
	helm install --dry-run --debug -f $(HELM_FILE) $(HELM_NAME) $(HELM_REPO)/$(HELM_CHART)

.PHONY: delete
delete:
	-helm delete $(HELM_NAME)


###############################################################################
###                                 Port forward                            ###
###############################################################################

# run command with `$ make port-forward FILE=configs/osmo-juno-cosmos.yaml`
.PHONY: port-forward
port-forward:
	bash $(CURDIR)/scripts/port-forward.sh --config=$(HELM_FILE)

.PHONY: stop-forward
stop-forward:
	-pkill -f "port-forward"

###############################################################################
###                          Local Kind Setup                               ###
###############################################################################
KIND_CLUSTER=starship

.PHONY: setup-kind
setup-kind:
	kind create cluster --name $(KIND_CLUSTER)

.PHONY: clean-kind
clean-kind:
	kind delete cluster --name $(KIND_CLUSTER)


