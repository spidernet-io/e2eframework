// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	spiderv2beta1 "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Spiderpool Multus", Label("spidermultus"), func() {

	var name, namespace string
	var f *e2e.Framework
	BeforeEach(func() {
		f = fakeFramework()
		name = "test"
		namespace = "kube-system"
	})

	It("Operate Spiderpool Multus Instance", func() {
		var err error
		err = f.CreateSpiderMultusInstance(&spiderv2beta1.SpiderMultusConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		})

		Expect(err).To(BeNil())
		nad, err := f.GetSpiderMultusInstance(namespace, name)
		Expect(err).NotTo(HaveOccurred())
		Expect(nad).NotTo(BeNil())

		list, err := f.ListSpiderMultusInstances()
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Printf("len of instances: %v", len(list.Items))

		err = f.DeleteSpiderMultusInstance(namespace, name)
		Expect(err).NotTo(HaveOccurred())

	})
})
