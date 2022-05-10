// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
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
			Replicas: pointer.Int32Ptr(replica),
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
		Status: appsv1.StatefulSetStatus{
			Replicas:      replica,
			ReadyReplicas: replica,
		},
	}
}

var _ = Describe("test statefulSet", Label("statefulSet"), func() {
	var f *e2e.Framework

	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate statefulSet", func() {

		stsName := "testSts"
		errStsName := "errName"
		namespace := "default"
		replica := int32(3)

		go func() {
			defer GinkgoRecover()
			// notice: WaitPodStarted use watch , but for the fake clientset,
			// the watch have started before the pod ready, or else the watch will miss the event
			// so we create the pod after WaitPodStarted
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			sts1 := generateExampleStatefulSetYaml(stsName, namespace, replica)
			e := f.CreateStatefulSet(sts1)
			Expect(e).NotTo(HaveOccurred())
			GinkgoWriter.Printf("finish creating statefulSet \n")

			// UT cover create the same deployment name
			err1 := f.CreateStatefulSet(sts1)
			Expect(err1).To(HaveOccurred())
			GinkgoWriter.Printf("failed to create , a same deployment %v/%v exists \n", namespace, stsName)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		sts, e1 := f.WaitStatefulSetReady(stsName, namespace, ctx)
		Expect(e1).NotTo(HaveOccurred())

		getSts, e2 := f.GetStatefulSet(stsName, namespace)
		Expect(e2).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get sts: %+v \n", getSts)

		// UT cover deployment name does not exist
		_, err21 := f.GetStatefulSet(errStsName, namespace)
		Expect(err21).To(HaveOccurred())
		GinkgoWriter.Printf("The unit test coverage name:%v does not exist", errStsName)

		pods, e3 := f.GetStatefulSetPodList(sts)
		Expect(e3).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get statefulSet podList: %v\n", pods)

		e4 := f.DeleteStatefulSet(stsName, namespace)
		Expect(e4).NotTo(HaveOccurred())
	})
})
