# kube2consul

A daemon that watches services in a Kubernetes cluster and registers them with Consul. If a service ceases to exist, kube2consul also deregisters the service from Consul.

## How to run kube2consul?
1. Clone the source code into your Go workspace
   ```
   $ git clone https://github.com/VeritasOS/kube2consul.git
   ```

2. Fetch dependencies
   ```
   $ dep ensure
   ```

   If `dep` is not installed, please install it using `go get -u github.com/golang/dep/cmd/dep`

3. Compile kube2consul
   ```
   $ go build
   ```

3. Run it!
   ```
   $ export K2C_DEBUG=true
   $ export KUBERNETES_SERVICE_HOST="k8s-api-server.example.com" # hostname of Kubernetes API server to connect to
   $ export KUBERNETES_SERVICE_PORT="443"                        # port number on which Kubernetes API server is listening
   $ export CONSUL_HTTP_ADDR="consul.example.com"                # hostname of Consul server to connect to
   $ export CONSUL_HTTP_SSL="true"                               # whether Consul is running with SSL enabled
   $ export CLOUD_PROVIDER="aws"                                 # cloud environment on which the Kubernetes cluster is running (supported values are "aws" or "openstack")

   $ ./kube2consul -kubeconfig /tmp/k8s.config                   # /tmp/k8s.config is the KUBECONFIG file for the Kubernetes cluster to watch
   ```

## Docker image
We plan to provide a `kube2consul` docker image that can be deployed to a Kubernetes cluster to start registering services from that cluster to Consul.

## Future enhancements
1. Use service account token instead of mounting client certificates

2. Use callback mechanism instead of polling Kubernetes periodically for service status
