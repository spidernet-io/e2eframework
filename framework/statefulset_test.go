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
	"k8s.io/utils/pointer"
)

func generateExampleStatefulSetYaml(stsName, namespace string, replica int32) *appsv1.StatefulSet {
	Expect(stsName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      stsName,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: pointer.Int32(replica),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": stsName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": stsName,
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
		//the fake clientset will not schedule statefulSet replicaset,so mock the number
		Status: appsv1.StatefulSetStatus{
			Replicas:        replica,
			ReadyReplicas:   replica,
			CurrentReplicas: replica,
		},
	}
}

var _ = Describe("test statefulSet", Label("statefulSet"), func() {
	var f *e2e.Framework
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate statefulSet", func() {
		stsName := "testSts"
		namespace := "default"
		replica := int32(3)
		scaleReplicas := int32(2)

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			time.Sleep(2 * time.Second)
			sts := generateExampleStatefulSetYaml(stsName, namespace, replica)

			// create statefulSet
			GinkgoWriter.Println("create statefulSet")
			e1 := f.CreateStatefulSet(sts)
			Expect(e1).NotTo(HaveOccurred())
			GinkgoWriter.Println("finish creating statefulSet")

			wg.Done()
		}()

		// wait statefulSet ready
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		GinkgoWriter.Println("wait for statefulSet ready")
		sts, e1 := f.WaitStatefulSetReady(stsName, namespace, ctx1)
		Expect(e1).NotTo(HaveOccurred())
		Expect(sts).NotTo(BeNil())
		GinkgoWriter.Println("statefulSet is ready")

		wg.Wait()

		// get statefulSet
		GinkgoWriter.Println("get statefulSet")
		getSts2, e2 := f.GetStatefulSet(stsName, namespace)
		Expect(e2).NotTo(HaveOccurred())
		Expect(getSts2).NotTo(BeNil())
		GinkgoWriter.Printf("get statefulSet: %v/%v \n", namespace, stsName)

		// get statefulSet pod list
		GinkgoWriter.Println("get statefulSet pod list")
		podList3, e3 := f.GetStatefulSetPodList(sts)
		Expect(podList3).NotTo(BeNil())
		Expect(e3).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get statefulSet podList: %+v \n", *podList3)

		// scale statefulSet
		GinkgoWriter.Println("scale statefulSet")
		getSts4, e4 := f.ScaleStatefulSet(sts, scaleReplicas)
		Expect(getSts4).NotTo(BeNil())
		Expect(e4).NotTo(HaveOccurred())
		GinkgoWriter.Printf("succeed to scale statefulSet %v to %v \n", stsName, scaleReplicas)

		// create a statefulSet with a same name
		GinkgoWriter.Println("create a statefulSet with a same name")
		time.Sleep(5 * time.Second)
		e5 := f.CreateStatefulSet(sts)
		Expect(e5).To(HaveOccurred())
		GinkgoWriter.Printf("failed creating a statefulSet with a same name: %v\n", stsName)

		// here start the operation of creating the statefulSet being deleted
		// cannot create a statefulSet being deleted
		GinkgoWriter.Printf("delete statefulSet %v \n", stsName)
		e6 := f.DeleteStatefulSet(stsName, namespace)
		Expect(e6).NotTo(HaveOccurred())
		GinkgoWriter.Printf("%v deleted successfully \n", stsName)

		// create a statefulSet being deleted
		GinkgoWriter.Println("create a statefulSet being deleted")
		e7 := f.CreateStatefulSet(sts)
		Expect(e7).To(HaveOccurred())
		GinkgoWriter.Printf("failed to create %v that being deleted\n", stsName)
	})
	It("counter example with wrong input", func() {
		stsName := "testSts"
		namespace := "default"
		scaleReplicas := int32(2)
		var stsNil *appsv1.StatefulSet = nil

		// failed wait statefulSet ready with wrong input
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		GinkgoWriter.Println("failed wait statefulSet ready with wrong input")
		sts2, e2 := f.WaitStatefulSetReady("", namespace, ctx1)
		Expect(sts2).To(BeNil())
		Expect(e2).Should(MatchError(e2e.ErrWrongInput))

		sts2, e2 = f.WaitStatefulSetReady(stsName, "", ctx1)
		Expect(sts2).To(BeNil())
		Expect(e2).Should(MatchError(e2e.ErrWrongInput))

		// failed to delete statefulSet with wrong input
		GinkgoWriter.Println("failed to delete statefulSet with wrong input")
		e3 := f.DeleteStatefulSet("", namespace)
		Expect(e3).Should(MatchError(e2e.ErrWrongInput))
		e3 = f.DeleteStatefulSet(stsName, "")
		Expect(e3).Should(MatchError(e2e.ErrWrongInput))

		// failed to get statefulSet with wrong input
		GinkgoWriter.Println("failed to get statefulSet with wrong input")
		getSts4, e4 := f.GetStatefulSet("", namespace)
		Expect(getSts4).To(BeNil())
		Expect(e4).Should(MatchError(e2e.ErrWrongInput))
		getSts4, e4 = f.GetStatefulSet(stsName, "")
		Expect(getSts4).To(BeNil())
		Expect(e4).Should(MatchError(e2e.ErrWrongInput))

		// failed to get statefulSet pod list with wrong input
		GinkgoWriter.Println("failed to get statefulSet pod list with wrong input")
		podList5, e5 := f.GetStatefulSetPodList(stsNil)
		Expect(podList5).To(BeNil())
		Expect(e5).Should(MatchError(e2e.ErrWrongInput))

		// failed to scale statefulSet with wrong input
		GinkgoWriter.Println("failed to scale statefulSet with wrong input")
		getSts6, e6 := f.ScaleStatefulSet(stsNil, scaleReplicas)
		Expect(getSts6).To(BeNil())
		Expect(e6).Should(MatchError(e2e.ErrWrongInput))
	})
})
