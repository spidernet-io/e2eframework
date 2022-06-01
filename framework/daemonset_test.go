// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateExampleDaemonSetYaml(dsName, namespace string, numberReady, desiredNumberScheduled int32) *appsv1.DaemonSet {
	Expect(dsName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      dsName,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dsName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": dsName,
					},
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
			},
		},
		// the fake clientset will not schedule daemonset replicaset， so mock the number
		Status: appsv1.DaemonSetStatus{
			NumberReady:            numberReady,
			DesiredNumberScheduled: desiredNumberScheduled,
		},
	}
}

var _ = Describe("unit test DaemonSet", Label("DaemonSet"), func() {
	var f *e2e.Framework
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate DaemonSet", func() {
		dsName := "testDs"
		namespace := "ds-ns"
		numberReady := int32(1)
		desiredNumberScheduled := int32(1)

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			time.Sleep(2 * time.Second)
			ds := generateExampleDaemonSetYaml(dsName, namespace, numberReady, desiredNumberScheduled)

			// create DaemonSet
			GinkgoWriter.Println("create DaemonSet")
			e1 := f.CreateDaemonSet(ds)
			Expect(e1).NotTo(HaveOccurred())
			GinkgoWriter.Println("finish creating DaemonSet")

			wg.Done()
		}()

		// wait DaemonSet ready
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		GinkgoWriter.Println("wait for DaemonSet ready")
		ds, e1 := f.WaitDaemonSetReady(dsName, namespace, ctx1)
		Expect(e1).NotTo(HaveOccurred())
		Expect(ds).NotTo(BeNil())
		GinkgoWriter.Println("DaemonSet is ready")

		wg.Wait()

		// get DaemonSet
		GinkgoWriter.Println("get DaemonSet")
		getds2, e2 := f.GetDaemonSet(dsName, namespace)
		Expect(e2).NotTo(HaveOccurred())
		Expect(getds2).NotTo(BeNil())
		GinkgoWriter.Printf("get DaemonSet: %v/%v \n", namespace, dsName)

		// get DaemonSet pod list
		GinkgoWriter.Println("get DaemonSet pod list")
		podList3, e3 := f.GetDaemonSetPodList(ds)
		Expect(podList3).NotTo(BeNil())
		Expect(e3).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get DaemonSet podList: %+v \n", *podList3)

		// create a DaemonSet with a same name
		GinkgoWriter.Println("create a DaemonSet with a same name")
		e5 := f.CreateDaemonSet(ds)
		Expect(e5).To(HaveOccurred())
		GinkgoWriter.Printf("failed creating a DaemonSet with a same name: %v\n", dsName)

		// delete already created daemonset
		GinkgoWriter.Printf("delete DaemonSet %v \n", dsName)
		e6 := f.DeleteDaemonSet(dsName, namespace)
		Expect(e6).NotTo(HaveOccurred())
		GinkgoWriter.Printf("%v deleted successfully \n", dsName)

	})
	It("counter example with wrong input", func() {
		dsName := "testDs"
		namespace := "ns-ds"
		var dsNil *appsv1.DaemonSet = nil

		// failed wait DaemonSet ready with wrong input
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		GinkgoWriter.Println("failed wait DaemonSet ready with wrong input")
		ds2, e2 := f.WaitDaemonSetReady("", namespace, ctx1)
		Expect(ds2).To(BeNil())
		Expect(e2).Should(MatchError(e2e.ErrWrongInput))

		ds2, e2 = f.WaitDaemonSetReady(dsName, "", ctx1)
		Expect(ds2).To(BeNil())
		Expect(e2).Should(MatchError(e2e.ErrWrongInput))

		// failed to delete DaemonSet with wrong input
		GinkgoWriter.Println("failed to delete DaemonSet with wrong input")
		e3 := f.DeleteDaemonSet("", namespace)
		Expect(e3).Should(MatchError(e2e.ErrWrongInput))
		e3 = f.DeleteDaemonSet(dsName, "")
		Expect(e3).Should(MatchError(e2e.ErrWrongInput))

		// failed to get DaemonSet with wrong input
		GinkgoWriter.Println("failed to get DaemonSet with wrong input")
		getds4, e4 := f.GetDaemonSet("", namespace)
		Expect(getds4).To(BeNil())
		Expect(e4).Should(MatchError(e2e.ErrWrongInput))
		getds4, e4 = f.GetDaemonSet(dsName, "")
		Expect(getds4).To(BeNil())
		Expect(e4).Should(MatchError(e2e.ErrWrongInput))

		// failed to get DaemonSet pod list with wrong input
		GinkgoWriter.Println("failed to get DaemonSet pod list with wrong input")
		podList5, e5 := f.GetDaemonSetPodList(dsNil)
		Expect(podList5).To(BeNil())
		Expect(e5).Should(MatchError(e2e.ErrWrongInput))

	})
})
