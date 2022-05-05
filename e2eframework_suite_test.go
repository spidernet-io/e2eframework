// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package e2eframework_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2eframework(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2eframework Suite")
}
