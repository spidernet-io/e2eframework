// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("test events", Label("events"), func() {
	var f *e2e.Framework
	var podName, namespace string
	var label map[string]string
	BeforeEach(func() {
		f = fakeFramework()

		podName = "testevents"
		namespace = "default"
		label = map[string]string{
			"app": "testevents",
		}
	})

	It("operate events ", func() {
		// generate pod yaml
		pod := generateExamplePodYaml(podName, namespace, label, "")

		// Constructing events by creating pods
		e := f.CreatePod(pod)
		Expect(e).NotTo(HaveOccurred())
		GinkgoWriter.Printf("finish creating pod %v/%v \n", namespace, podName)

		// get events
		events, e2 := f.ListEvents(&client.ListOptions{
			Raw: &metav1.ListOptions{
				TypeMeta:      metav1.TypeMeta{Kind: "Pod"},
				FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", podName, namespace),
			},
		})
		Expect(e2).NotTo(HaveOccurred())
		Expect(events).NotTo(BeNil())

		// delete pod
		e3 := f.DeletePod(podName, namespace)
		Expect(e3).NotTo(HaveOccurred())
	})
})
