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

// Mechanism used to talk to the Kubernetes cluster is inspired from
// https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster/main.go

package main

import (
	"flag"
	"os"
	"github.com/VeritasOS/kube2consul/client"
	"github.com/VeritasOS/kube2consul/logging"
	"strconv"
	"time"
)

func main() {
	// parse command-line arguments
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	isDebug := flag.Bool("debug", false, "verbose logs")
	flag.Parse()

	if *isDebug == false {
		envValue := os.Getenv("K2C_DEBUG")
		if envValue != "" {
			boolValue, err := strconv.ParseBool(envValue)
			if err != nil {
				panic(err.Error())
			}
			isDebug = &boolValue
		}
	}

	logging.InitLogger(isDebug)
	var log = logging.GetInstance()

	kubeServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	if kubeServiceHost == "" {
		log.Fatal("KUBERNETES_SERVICE_HOST not set")
		os.Exit(1)
	}
	cloudProvider := os.Getenv("CLOUD_PROVIDER")
	if cloudProvider == "" {
		log.Fatal("CLOUD_PROVIDER not set")
		os.Exit(1)
	}

	var kubernetesClient *client.KubernetesClient = nil
	if *kubeconfig == "" {
		kubernetesClient = client.InternalKubernetesClient()
	} else {
		kubernetesClient = client.ExternalKubernetesClient(kubeconfig)
	}

	consulClient := client.NewConsulClient()

	log.Debug("Entering main loop")
	for {
		// get services in the Kubernetes cluster
		kubeServices := kubernetesClient.Services()

		// register new services with Consul
		for _, kubeService := range kubeServices {

			if kubernetesClient.IsSystemService(kubeService) {
				continue
			}

			kubeServiceName := kubeService.ObjectMeta.Name
			kubeServiceNamespace := kubeService.ObjectMeta.Namespace

			if !consulClient.IsRegistered(kubeServiceName, kubeServiceNamespace, kubeServiceHost) {
				// register new service
				endpoints := kubernetesClient.ServiceEndpoints(kubeService, cloudProvider)
				log.Debugf("Endpoins for service [%s][%s][%s] are [%#v]", kubeServiceName, kubeServiceNamespace, kubeServiceHost, endpoints)
				if len(endpoints) > 0 {
					log.Infof("Registering service [%s][%s][%s] with endpoints [%#v]", kubeServiceName, kubeServiceNamespace, kubeServiceHost, endpoints)
					consulClient.Register(kubeServiceName, kubeServiceNamespace, kubeServiceHost, endpoints)
				}
			}

		}

		// get services registered with Consul
		consulServices := consulClient.Services()

		// deregister non-existent services from consul
		for _, consulService := range consulServices {

			if consulClient.IsSystemService(consulService) {
				continue
			}

			kubeServiceNameTag, kubeServiceNamespaceTag, kubeServiceHostTag := consulClient.ParseTags(consulService)

			// Check if consul service belongs to "this" cluster
			if kubeServiceHost == kubeServiceHostTag {
				if !kubernetesClient.IsRunning(kubeServiceNameTag, kubeServiceNamespaceTag) {
					// deregister service
					id := consulService.ID
					log.Infof("Deregistering service [%s][%s][%s] (ID=[%s])", kubeServiceNameTag, kubeServiceNamespaceTag, kubeServiceHostTag, id)
					consulClient.Deregister(id)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
