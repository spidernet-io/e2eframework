// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spidernet-io/e2eframework/framework"
	"time"
)

var _ = Describe("Command", Label("Command"), func() {
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

	It("docker command", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// create container
		var name = "test-docker-command"
		command := fmt.Sprintf("-itd --name %s busybox", name)
		containerId, err := f.DockerRunCommand(ctx, command)
		Expect(err).NotTo(HaveOccurred(), "failed to run 'docker run': %v", string(containerId))
		Expect(containerId).NotTo(BeNil())

		output, err := f.DockerExecCommand(ctx, name, "echo hello world")
		Expect(err).NotTo(HaveOccurred(), "failed to run 'docker exec': %v", string(output))
		Expect(output).NotTo(BeNil())
		GinkgoWriter.Printf("output: %s", string(output))

		output, err = f.DockerRMCommand(ctx, name)
		Expect(err).NotTo(HaveOccurred(), "failed to run 'docker rm -f': %v", string(output))

	})
})
