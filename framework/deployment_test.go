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
	"k8s.io/utils/pointer"
)

func GenerateExampleDeploymentYaml(dpmName, namespace string, replica, readyReplica int32) *appsv1.Deployment {
	Expect(dpmName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      dpmName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(replica),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dpmName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": dpmName,
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
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: readyReplica,
			Replicas:      replica,
		},
	}
}

var _ = Describe("test deployment", Label("deployment"), func() {
	var f *e2e.Framework

	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate deployment", func() {
		dpmName := "testDpm"
		errName := "errDpm"
		namespace := "default"
		replica := int32(3)
		readyReplica := int32(3)

		go func() {
			defer GinkgoRecover()
			// notice: WaitPodStarted use watch , but for the fake clientset,
			// the watch have started before the pod ready, or else the watch will miss the event
			// so we create the pod after WaitPodStarted
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			dpm := GenerateExampleDeploymentYaml(dpmName, namespace, replica, readyReplica)
			err1 := f.CreateDeployment(dpm)
			Expect(err1).NotTo(HaveOccurred())
			GinkgoWriter.Printf("finish creating deployment \n")

			// UT cover create the same deployment name
			dpm = GenerateExampleDeploymentYaml(dpmName, namespace, replica, readyReplica)
			err2 := f.CreateDeployment(dpm)
			Expect(err2).To(HaveOccurred())
			GinkgoWriter.Printf("failed to create , a same deployment default/testDpm exists \n")
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()
		_, err3 := f.WaitDeploymentReady(dpmName, namespace, ctx)
		Expect(err3).NotTo(HaveOccurred())

		getDpm, err4 := f.GetDeploymnet(dpmName, namespace)
		Expect(err4).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get deployment: %+v \n", getDpm)

		// UT cover deployment name does not exist
		_, err4 = f.GetDeploymnet(errName, namespace)
		Expect(err4).To(HaveOccurred())
		GinkgoWriter.Printf("The unit test coverage name:%v does not exist", errName)

		err5 := f.DeleteDeployment(dpmName, namespace)
		Expect(err5).NotTo(HaveOccurred())
	})

})
