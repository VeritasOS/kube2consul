/*
Copyright 2018 Veritas Technologies LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	clientset *kubernetes.Clientset
}

func InternalKubernetesClient() *KubernetesClient {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	config.CertFile = "/root/kube2consul/certs/cert.pem"
	config.KeyFile = "/root/kube2consul/certs/key.pem"
	config.CAFile = "/root/kube2consul/certs/ca-cert.pem"
	return newKubeClient(config)
}

func ExternalKubernetesClient(kubeconfig *string) *KubernetesClient {
	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return newKubeClient(config)
}

func newKubeClient(kubeconfig *rest.Config) *KubernetesClient {
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	instance := &KubernetesClient{clientset}
	return instance
}

func (kubernetesClient *KubernetesClient) Services() []v1.Service {
	kubeServices, err := kubernetesClient.clientset.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	return kubeServices.Items
}

func (kubernetesClient *KubernetesClient) Nodes() []v1.Node {
	kubeNodes, err := kubernetesClient.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	return kubeNodes.Items
}

func (kubernetesClient *KubernetesClient) IsSystemService(kubeService v1.Service) bool {
	// skip system services
	if kubeService.ObjectMeta.Namespace == "kube-system" {
		return true
	}

	// skip default kubernetes services
	if kubeService.ObjectMeta.Namespace == "default" && kubeService.ObjectMeta.Name == "kubernetes" {
		return true
	}

	// skip consul services
	if kubeService.ObjectMeta.Namespace == "consul" {
		return true
	}

	return false
}

func (kubernetesClient *KubernetesClient) ServiceEndpoints(kubeService v1.Service, cloudProvider string) []ServiceEndpoint {

	var endpoints []ServiceEndpoint
	var addr string
	if kubeService.Spec.Type == "LoadBalancer" {
		// Register only after external LB IP is available
		if cloudProvider == "aws" {
			if len(kubeService.Status.LoadBalancer.Ingress) > 0 {
				addr = kubeService.Status.LoadBalancer.Ingress[0].Hostname
			}
		} else if cloudProvider == "openstack" {
			if len(kubeService.Status.LoadBalancer.Ingress) > 1 {
				addr = kubeService.Status.LoadBalancer.Ingress[1].IP
			}
		}
		if addr != "" {
			port := 80 // TODO: check if port can be something other than 80
			e := ServiceEndpoint{addr, port}
			endpoints = append(endpoints, e)
		}

	} else if kubeService.Spec.Type == "NodePort" {
		addresses := kubernetesClient.NodeAddresses()
		port := int(kubeService.Spec.Ports[0].NodePort)

		for _, addr := range addresses {
			e := ServiceEndpoint{addr, port}
			endpoints = append(endpoints, e)
		}

	} else if kubeService.Spec.Type == "ClusterIP" {
		addr := kubeService.Spec.ClusterIP
		if addr != "None" {
			port := int(kubeService.Spec.Ports[0].TargetPort.IntVal)
			e := ServiceEndpoint{addr, port}
			endpoints = append(endpoints, e)
		}
	}

	return endpoints
}

func (kubernetesClient *KubernetesClient) NodeAddresses() []string {
	kubeNodes := kubernetesClient.Nodes()

	var internalAddrList []string
	var externalAddrList []string
	for _, node := range kubeNodes {

		for _, addr := range node.Status.Addresses {
			if addr.Type == "ExternalIP" {
				externalAddrList = append(externalAddrList, addr.Address)
			}
			if addr.Type == "InternalIP" {
				internalAddrList = append(internalAddrList, addr.Address)
			}
		}
	}

	// If all nodes have external IP address use them
	if len(externalAddrList) == len(kubeNodes) {
		log.Debugf("Node list : [%v]", externalAddrList)
		return externalAddrList
	}

	log.Debugf("Node list : [%v]", externalAddrList)
	return internalAddrList
}

func (kubernetesClient *KubernetesClient) IsRunning(kubeServiceName, kubeServiceNamespace string) bool {
	kubeServices := kubernetesClient.Services()

	for _, kubeService := range kubeServices {
		if kubeService.ObjectMeta.Name == kubeServiceName && kubeService.ObjectMeta.Namespace == kubeServiceNamespace {
			log.Debugf("Service [%s] from namespace [%s] is running on Kubernetes", kubeServiceName, kubeServiceNamespace)
			return true
		}
	}

	return false
}
