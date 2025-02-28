export CGO_ENABLED = 0

all: build

fmt:
	go fmt ./...

vet:
	go vet ./... && go run honnef.co/go/tools/cmd/staticcheck@latest ./...

ifeq ($(OS),Windows_NT)
    DEV_NULL = NUL
else
    DEV_NULL = /dev/null
endif

test:
	go test -v ./... -coverprofile=$(DEV_NULL)

validate: fmt lint vet test

build: validate
	go build -o ./dist/service-resolver ./src

run:
	go run main.go

container:
	docker build -t service-resolver:latest .

cluster: clean
	kind create cluster --config test-cluster/manifests/kind/cluster.yml
	
deploy:
	kubectl apply -k test-cluster 
	kind load docker-image service-resolver:latest -n local
	helm upgrade --install -n hello-world service-resolver helm --set imagePullPolicy=Never --wait

port-forward:
	kubectl -n hello-world port-forward svc/service-resolver 8080

clean: 
	kind delete cluster -n local

.PHONY: all fmt lint vet test validate build run container cluster clean deploy