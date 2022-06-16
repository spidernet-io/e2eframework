// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateConfigmapYaml(name, namespace string) *corev1.ConfigMap {
	Expect(name).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

var _ = Describe("test configmap", Label("configmap"), func() {
	var (
		f               *e2e.Framework
		name, namespace string
	)

	BeforeEach(func() {
		f = fakeFramework()
		name = "test-cm"
		namespace = "test-ns"
	})

	It("operate configmap", func() {
		// create configmap
		configmapYaml := generateConfigmapYaml(name, namespace)
		GinkgoWriter.Printf("create configmap %v/%v\n", namespace, name)
		Expect(f.CreateConfigmap(configmapYaml)).To(Succeed())

		// get configmap
		GinkgoWriter.Printf("get configmap %v/%v\n", namespace, name)
		cm1, e1 := f.GetConfigmap(name, namespace)
		Expect(cm1).NotTo(BeNil())
		Expect(e1).NotTo(HaveOccurred())

		// create same-name configmap
		GinkgoWriter.Printf("create same-name configmap %v/%v\n", namespace, name)
		e2 := f.CreateConfigmap(configmapYaml)
		Expect(e2).To(HaveOccurred())

		// delete configmap
		GinkgoWriter.Printf("delete configmap %v/%v\n", namespace, name)
		Expect(f.DeleteConfigmap(name, namespace)).To(Succeed())

	})
	It("counter example with wrong input", func() {
		// create configmap with wrong input
		GinkgoWriter.Println("create configmap with wrong input")
		e1 := f.CreateConfigmap(nil)
		Expect(e1).Should(MatchError(e2e.ErrWrongInput))

		// get configmap with wrong input
		GinkgoWriter.Println("get config map with wrong input")
		cm2, e2 := f.GetConfigmap("", namespace)
		Expect(cm2).To(BeNil())
		Expect(e2).Should(MatchError(e2e.ErrWrongInput))

		cm3, e3 := f.GetConfigmap(name, "")
		Expect(cm3).To(BeNil())
		Expect(e3).Should(MatchError(e2e.ErrWrongInput))

		// delete configmap with wrong input
		GinkgoWriter.Println("delete configmap with wrong input")
		e4 := f.DeleteConfigmap("", namespace)
		Expect(e4).Should(MatchError(e2e.ErrWrongInput))

		e5 := f.DeleteConfigmap(name, "")
		Expect(e5).Should(MatchError(e2e.ErrWrongInput))
	})
})
