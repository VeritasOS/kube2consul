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
	"github.com/hashicorp/consul/api"
	"github.com/satori/go.uuid"
	"github.com/VeritasOS/kube2consul/utils"
	"strings"
)

type ConsulClient struct {
	agent *api.Agent
}

func NewConsulClient() *ConsulClient {

	// get a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	agent := client.Agent()

	return &ConsulClient{agent}
}

func (consulClient *ConsulClient) Services() map[string]*api.AgentService {
	services, err := consulClient.agent.Services()
	if err != nil {
		panic(err)
	}

	return services
}

func (consulClient *ConsulClient) IsSystemService(consulService *api.AgentService) bool {
	if consulService.Service == "consul" {
		return true
	}
	return false
}

func (consulClient *ConsulClient) IsRegistered(kubeServiceName, kubeServiceNamespace, kubeServiceHost string) bool {
	// TODO: Try to minimize calls to consulClient.Services()
	log.Debugf("Checking if [%s][%s][%s] is registered with Consul", kubeServiceName, kubeServiceNamespace, kubeServiceHost)

	name := strings.Join([]string{kubeServiceName, kubeServiceNamespace}, "-")
	tags := []string{"name=" + kubeServiceName, "ns=" + kubeServiceNamespace, "kube=" + kubeServiceHost}

	for _, consulService := range consulClient.Services() {
		if name == consulService.Service && utils.Subset(tags, consulService.Tags) {
			return true
		}
	}
	return false
}

func (consulClient *ConsulClient) Register(kubeServiceName, kubeServiceNamespace, kubeServiceHost string, endpoints []ServiceEndpoint) {

	for _, endpoint := range endpoints {

		id := uuid.NewV4().String()
		name := strings.Join([]string{kubeServiceName, kubeServiceNamespace}, "-")
		tags := []string{"name=" + kubeServiceName, "ns=" + kubeServiceNamespace, "kube=" + kubeServiceHost, "id=" + id}

		consulService := &api.AgentServiceRegistration{
			id,
			name,
			tags,
			endpoint.port,
			endpoint.addr,
			false,
			nil,
			make([]*api.AgentServiceCheck, 0),
		}

		consulClient.agent.ServiceRegister(consulService)
		log.Infof("Registered endpoint [%#v] for service [%s][%s][%s] with Consul (ID=[%s])", endpoint, kubeServiceName, kubeServiceNamespace, kubeServiceHost, id)
	}
}

func (consulClient *ConsulClient) ParseTags(consulService *api.AgentService) (kubeServiceNameTag, kubeServiceNamespaceTag, kubeServiceHostTag string) {

	for _, tag := range consulService.Tags {
		tokens := strings.Split(tag, "=")
		switch tokens[0] {
		case "name":
			kubeServiceNameTag = tokens[1]
		case "ns":
			kubeServiceNamespaceTag = tokens[1]
		case "kube":
			kubeServiceHostTag = tokens[1]
		}
	}

	return
}

func (consulClient *ConsulClient) Deregister(consulServiceId string) {
	consulClient.agent.ServiceDeregister(consulServiceId)
	log.Infof("Service with ID [%s] deregistered from Consul", consulServiceId)
}
