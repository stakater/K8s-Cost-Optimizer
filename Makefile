VERSION ?= 0.0.1
GOCMD = go
BINARY ?= k8s-cost-optimizer

test: fmt
	go test -v ./...
fmt:
	gofmt -s -w .

install:
	"$(GOCMD)" mod download

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "$(GOCMD)" build -o "${BINARY}"

helm-lint:
	if [ -d charts ]; then helm lint charts/kubehealth; fi

bump-chart-version: 
	sed -i "s/^version:.*/version: v$(VERSION)/" deployments/kubernetes/chart/k8s-cost-optimizer/Chart.yaml
	sed -i "s/^appVersion:.*/appVersion: v$(VERSION)/" deployments/kubernetes/chart/k8s-cost-optimizer/Chart.yaml
	sed -i "s/tag:.*/tag: v$(VERSION)/" deployments/kubernetes/chart/k8s-cost-optimizer/values.yaml
	sed -i "s/version:.*/version: v$(VERSION)/" deployments/kubernetes/chart/k8s-cost-optimizer/values.yaml

# Bump Chart
bump-chart: bump-chart-version