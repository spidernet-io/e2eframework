// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package e2eframework_test

import (
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
		e := f.CreateNamespace(namespace)
		Expect(e).NotTo(HaveOccurred())

		e = f.DeleteNamespace(namespace)
		Expect(e).NotTo(HaveOccurred())

	})

})
