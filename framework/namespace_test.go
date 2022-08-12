// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e2e "github.com/spidernet-io/e2eframework/framework"
)

var _ = Describe("test namespace", Label("namespace"), func() {
	var f *e2e.Framework

	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate namespace", func() {
		namespace := "test"
		// counter case: create namespace but default service account not ready
		GinkgoWriter.Printf("counter case: create namespace %v but default service account not ready\n", namespace)
		Expect(f.CreateNamespaceUntilDefaultServiceAccountReady(namespace, time.Second)).NotTo(Succeed())

		// delete namespace
		GinkgoWriter.Printf("delete namespace %v\n", namespace)
		Expect(f.DeleteNamespace(namespace)).To(Succeed())

		// create default service account while created namespace
		go func() {
			defer GinkgoRecover()
			// sleep 10 Millisecond to wait namespace created succeeded
			time.Sleep(time.Millisecond * 10)
			GinkgoWriter.Printf("create default service account while created namespace %v\n", namespace)
			serviceAccountObj := GenerateExampleServiceAccountObj("default", namespace)
			Expect(CreateServiceAccount(f, serviceAccountObj)).To(Succeed(), "failed to create default serviceAccount %v/default\n", namespace)
		}()

		// create namespace
		e := f.CreateNamespaceUntilDefaultServiceAccountReady(namespace, time.Second*5)
		Expect(e).NotTo(HaveOccurred())

		ns, e1 := f.GetNamespace(namespace)
		Expect(ns).NotTo(BeNil())
		Expect(e1).NotTo(HaveOccurred())

		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
		defer cancel1()
		e2 := f.DeleteNamespaceUntilFinish(namespace, ctx1)
		Expect(e2).NotTo(HaveOccurred())
	})
	It("counter example with wrong input", func() {
		// create namespace until default serviceAccount ready
		err := f.CreateNamespaceUntilDefaultServiceAccountReady("", time.Second)
		Expect(err).Should(MatchError(e2e.ErrWrongInput))

		e := f.DeleteNamespace("")
		Expect(e).Should(MatchError(e2e.ErrWrongInput))

		// delete namespace until finish, with wrong input
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Minute)
		defer cancel1()
		e1 := f.DeleteNamespaceUntilFinish("", ctx1)
		Expect(e1).To(HaveOccurred())
	})
})
