# service-resolver

This project provides API endpoints that can be called from outside a Kubernetes cluster to discover what IP addresses 
a given Kubernetes service is utilizing. If all of your workloads co-exist in a single k8s cluster you don't need this project.

Utilizing service-resolver in your cluster for long periods of time is probably a bad idea, but for periods of transition into
or out of Kubernetes it can help provide configuration properties to services at runtime.

## What is this for?
This project is most useful for helping to configure software running adjacent to a Kubernetes cluster that already has network 
level access to the cluster's resources but lack realtime service discovery or access to internal cluster DNS. Certain
implementations of architectures relying on AWS Lambda, Google Cloud Run, and Azure Functions can struggle with this issue.

## API Endpoints
This service provides the following API endpoints

### `v1/service/:namespace/:service`
Returns the k8s service level IP address for a given `:service`

### `v1/endpoints/:namespace/:service`
Returns the actual pod IP addresses for a given `:service` based on that service's selector configuration. This is useful
for many clustered technologies like Cassandra and Kafka and supporting the libraries that connect to them directly via IP

#### Local Dev
Dependencies
- make
- curl
- golang
- kind
- kubectl

```
# generate deployable artifact
make container

# turn on local cluster
make cluster

# ship test to local cluster
make deploy

# allow ingress
make port-forward

# test endpoints(open second-terminal)
## returns ip address
curl localhost:8080/v1/service/hello-world/home

## returns comma separated list of ip addresses
curl localhost:8080/v1/endpoints/hello-world/home

# shutdown ingress(return to first-terminal)
CTRL+C

# shut down the local test environment
make clean
```
