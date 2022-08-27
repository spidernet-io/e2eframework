// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spidernet-io/e2eframework/framework"
	"time"
)

var _ = Describe("Kind", Label("kind"), func() {
	var f *framework.Framework
	var err error
	BeforeEach(func() {
		f = fakeFramework()
	})
	It("exec kubectl", func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err = f.ExecKubectl("get po", ctx)
		Expect(err).To(HaveOccurred())
	})
})
