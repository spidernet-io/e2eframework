// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package e2eframework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e2e "github.com/spidernet-io/e2eframework"
)

var _ = Describe("test new Framework", Label("framework"), func() {

	DescribeTable("test environment",
		func(envlist map[string]string, expectedSucceed bool) {
			_, e := e2e.NewFramework(GinkgoT())
			if expectedSucceed == true {
				Expect(e).NotTo(HaveOccurred())
			} else {
				Expect(e).To(HaveOccurred())
			}
		},
		Entry("no env", nil, false),
		Entry("no env", map[string]string{
			e2e.E2E_CLUSTER_NAME: "testCluster",
		}, false),
	)
})
