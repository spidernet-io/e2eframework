// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func generateExampleEndpointYaml(name, namespace string, labels map[string]string) *v1.Endpoints {
	return &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "dummy_ip",
						NodeName: pointer.String("kind-test"),
						TargetRef: &v1.ObjectReference{
							Kind:      "Pod",
							Name:      name,
							Namespace: namespace,
						},
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     "http",
						Port:     80,
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}
}

var _ = Describe("Endpoint", Label("Endpoint"), func() {

	var f *e2e.Framework = fakeFramework()
	var name, namespace string
	var labels = map[string]string{
		"app": "test",
	}

	It("Operate Endpoint", func() {
		name = "test-ep"
		namespace = "default"
		var ep *v1.Endpoints
		var epList *v1.EndpointsList
		ept := generateExampleEndpointYaml(name, namespace, labels)

		// create test ep
		err := f.CreateEndpoint(ept)
		Expect(err).NotTo(HaveOccurred())

		// get test ep
		ep, err = f.GetEndpoint(name, namespace)
		Expect(err).NotTo(HaveOccurred())
		Expect(ep).NotTo(BeNil())

		// list eps
		epList, err = f.ListEndpoint()
		Expect(err).NotTo(HaveOccurred())
		Expect(epList).NotTo(BeNil())

		// delete ep
		err = f.DeleteEndpoint(name, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
})
