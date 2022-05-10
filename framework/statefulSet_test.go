// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"

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
	}
}

var _ = Describe("test pod", Label("pod"), func() {
	var f *e2e.Framework

	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate statefulSet", func() {

		stsName := "testSts"
		namespace := "default"
		replica := int32(3)
		sts := generateExampleStatefulSetYaml(stsName, namespace, replica)
		e := f.CreateStatefulSet(sts)
		Expect(e).NotTo(HaveOccurred())

		getSts, e1 := f.GetStatefulSet(stsName, namespace)
		Expect(e1).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get pod: %+v \n", getSts)

		e = f.DeleteStatefulSet(stsName, namespace)
		Expect(e).NotTo(HaveOccurred())
	})
})
