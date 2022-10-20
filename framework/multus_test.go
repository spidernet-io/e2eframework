// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	v1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Multus", Label("Multus"), func() {

	var name, namespace string
	var f *e2e.Framework
	BeforeEach(func() {
		f = fakeFramework()
		name = "test"
		namespace = "default"
	})

	It("Operate Multus Instance", func() {
		var err error
		err = f.CreateMultusInstance(&v1.NetworkAttachmentDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		})

		Expect(err).To(BeNil())
		nad, err := f.GetMultusInstance(name, namespace)
		Expect(err).NotTo(HaveOccurred())
		Expect(nad).NotTo(BeNil())

		list, err := f.ListMultusInstances()
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Printf("len of instances: %v", len(list.Items))

		err = f.DeleteMultusInstance(name, namespace)
		Expect(err).NotTo(HaveOccurred())

	})
})
