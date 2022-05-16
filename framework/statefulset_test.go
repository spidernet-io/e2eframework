// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e2e "github.com/spidernet-io/e2eframework/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": stsName},
			},
			Template: corev1.PodTemplateSpec{
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
	}
}

var _ = Describe("test statefulSet", Label("statefulSet"), func() {
	var f *e2e.Framework

	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate statefulSet", func() {

		stsName := "testSts"
		namespace := "default"
		replica := int32(3)
		sts := generateExampleStatefulSetYaml(stsName, namespace, replica)

		go func() {
			defer GinkgoRecover()

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			defer cancel()
			_, err := f.WaitStatefulSetReady(stsName, namespace, ctx)
			Expect(err).NotTo(HaveOccurred())
		}()

		e := f.CreateStatefulSet(sts)
		Expect(e).NotTo(HaveOccurred())

		getSts, e1 := f.GetStatefulSet(stsName, namespace)
		Expect(e1).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get sts: %+v \n", getSts)

		pods, e := f.GetStatefulSetPodList(sts)
		Expect(e).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get statefulSet podList: %v\n", pods)

		e = f.DeleteStatefulSet(stsName, namespace)
		Expect(e).NotTo(HaveOccurred())
	})
})
