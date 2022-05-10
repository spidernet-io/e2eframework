// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e2e "github.com/spidernet-io/e2eframework/framework"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateExamplePodYaml(podName, namespace string) *corev1.Pod {
	Expect(podName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      podName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "samplepod",
					Image:           "alpine",
					ImagePullPolicy: "IfNotPresent",
					Command:         []string{"/bin/ash", "-c", "trap : TERM INT; sleep infinity & wait"},
				},
			},
		},
	}
}

var _ = Describe("test pod", Label("pod"), func() {
	var f *e2e.Framework

	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate pod", func() {

		podName := "testpod"
		namespace := "default"

		go func() {
			defer GinkgoRecover()
			// notice: WaitPodStarted use watch , but for the fake clientset,
			// the watch have started before the pod ready, or else the watch will miss the event
			// so we create the pod after WaitPodStarted
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			pod1 := generateExamplePodYaml(podName, namespace)
			e2 := f.CreatePod(pod1)
			Expect(e2).NotTo(HaveOccurred())
			GinkgoWriter.Printf("finish creating pod \n")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		pod, e1 := f.WaitPodStarted(podName, namespace, ctx)
		Expect(e1).NotTo(HaveOccurred())
		Expect(pod).NotTo(BeNil())

		getPod, e1 := f.GetPod(podName, namespace)
		Expect(e1).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get pod: %+v \n", getPod)

		pods, e2 := f.GetPodList(&client.ListOptions{Namespace: namespace})
		Expect(e2).NotTo(HaveOccurred())
		GinkgoWriter.Printf("len of pods: %v", len(pods.Items))

		e := f.DeletePod(podName, namespace)
		Expect(e).NotTo(HaveOccurred())

	})

})
