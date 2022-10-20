// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateExampleServiceYaml(name, namespace string, labels map[string]string, port int32) *v1.Service {
	ipFamilyPolicy := v1.IPFamilyPolicyPreferDualStack
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			Ports: []v1.ServicePort{
				{
					Protocol: v1.ProtocolTCP,
					Port:     port,
				},
			},
			IPFamilyPolicy: &ipFamilyPolicy,
		},
	}
}

var _ = Describe("Service", Label("Service"), func() {
	var f *e2e.Framework
	var svcName, namespace string
	var label map[string]string

	BeforeEach(func() {
		f = fakeFramework()

		svcName = "test-svc"
		namespace = "default"
		label = map[string]string{
			svcName: svcName,
		}
	})
	It("operate service", func() {
		// generate example service yaml
		GinkgoWriter.Printf("create service %s/%s \n", namespace, svcName)
		serviceYaml := generateExampleServiceYaml(svcName, namespace, label, 80)
		Expect(serviceYaml).NotTo(BeNil(), "failed to generateExampleServiceYaml\n")

		err := f.CreateService(serviceYaml)
		Expect(err).NotTo(HaveOccurred(), "failed to CreateService, details: %v\n", err)

		svc, err := f.GetService(svcName, namespace)
		Expect(err).NotTo(HaveOccurred())
		Expect(svc).NotTo(BeNil())

		svcList, err := f.ListService()
		Expect(err).NotTo(HaveOccurred())
		Expect(svcList).NotTo(BeNil())

		err = f.DeleteService(svcName, namespace)
		Expect(err).NotTo(HaveOccurred())

	})
	It("counter example with wrong input", func() {
		// creat service with invalid input
		err := f.CreateService(nil)
		Expect(err).To(MatchError(e2e.ErrWrongInput))

		// get service with invalid input
		service, err := f.GetService("", namespace)
		Expect(err).To(MatchError(e2e.ErrWrongInput))
		Expect(service).To(BeNil())
		service, err = f.GetService(svcName, "")
		Expect(err).To(MatchError(e2e.ErrWrongInput))
		Expect(service).To(BeNil())

	})
})
