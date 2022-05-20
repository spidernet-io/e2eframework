// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"

	e2e "github.com/spidernet-io/e2eframework/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("test new Framework", Label("framework"), func() {
	var kubeConfigFile *os.File
	var fakeClient client.WithWatch

	BeforeEach(func() {
		kubeConfigFile = fakeKubeConfig()
		fakeEnv(kubeConfigFile.Name())
		fakeClient = fakeClientSet()
		Expect(fakeClient).NotTo(BeNil())

		DeferCleanup(func() {
			os.Remove(kubeConfigFile.Name())
			clearEnv()
		})
	})

	It("create framework", func() {
		_, e := e2e.NewFramework(GinkgoT(), fakeClient)
		Expect(e).NotTo(HaveOccurred())
	})

})
