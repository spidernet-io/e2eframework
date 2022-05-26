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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func generateExampleJobYaml(jbName, namespace string, parallelism *int32) *batchv1.Job {
	Expect(jbName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      jbName,
		},
		Spec: batchv1.JobSpec{

			Parallelism: parallelism,

			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": jbName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": jbName,
					},
				},
				Spec: corev1.PodSpec{
					//RestartPolicy: "Always",
					Containers: []corev1.Container{
						{
							Name:            "samplepod",
							Image:           "alpine",
							ImagePullPolicy: "IfNotPresent",
							Command:         []string{"/bin/ash", "-c", "trap : TERM INT; sleep 30 infinity & wait"},
						},
					},
				},
			},
		},
		// the fake clientset will not schedule Job  so mock the number

		// Status: batchv1.JobStatus{

		// 	Ready: ready,
		// },
	}
}

var _ = Describe("unit test Job", Label("Job"), func() {
	var f *e2e.Framework
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate Job", func() {
		jbName := "testjb"
		namespace := "ns-jb"
		//Ready := pointer.Int32(3)
		Parallelism := pointer.Int32(1)

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			//	time.Sleep(2 * time.Second)
			jb := generateExampleJobYaml(jbName, namespace, Parallelism)

			// create Job
			GinkgoWriter.Println("create Job")
			e1 := f.CreateJob(jb)
			Expect(e1).NotTo(HaveOccurred())
			GinkgoWriter.Println("finish creating Job")

			wg.Done()
		}()

		// wait Job ready
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		GinkgoWriter.Println("wait for Job ready")
		jb, e1 := f.WaitJobReady(jbName, namespace, ctx1)
		Expect(e1).NotTo(HaveOccurred())
		Expect(jb).NotTo(BeNil())
		GinkgoWriter.Println("Job is ready")

		wg.Wait()

		// get Job
		GinkgoWriter.Println("get Job")
		getjb2, e2 := f.GetJob(jbName, namespace)
		Expect(e2).NotTo(HaveOccurred())
		Expect(getjb2).NotTo(BeNil())
		GinkgoWriter.Printf("get Job: %v/%v \n", namespace, jbName)

		// get Job pod list
		GinkgoWriter.Println("get Job pod list")
		podList3, e3 := f.GetJobPodList(jb)
		Expect(podList3).NotTo(BeNil())
		Expect(e3).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get Job podList: %+v \n", *podList3)

		// create a Job with a same name
		GinkgoWriter.Println("create a Job with a same name")
		e5 := f.CreateJob(jb)
		Expect(e5).To(HaveOccurred())
		GinkgoWriter.Printf("failed creating a Job with a same name: %v\n", jbName)

		// delete already created Job
		GinkgoWriter.Printf("delete Job %v \n", jbName)
		e6 := f.DeleteJob(jbName, namespace)
		Expect(e6).NotTo(HaveOccurred())
		GinkgoWriter.Printf("%v deleted successfully \n", jbName)

	})
	It("counter example with wrong input", func() {
		jbName := "testjb"
		namespace := "ns-jb"
		var jbNil *batchv1.Job = nil

		// failed wait Job ready with wrong input
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		GinkgoWriter.Println("failed wait Job ready with wrong input")
		jb2, e2 := f.WaitJobReady("", namespace, ctx1)
		Expect(jb2).To(BeNil())
		Expect(e2).To(HaveOccurred())

		jb2, e2 = f.WaitJobReady(jbName, "", ctx1)
		Expect(jb2).To(BeNil())
		Expect(e2).To(HaveOccurred())

		// failed to delete Job with wrong input
		GinkgoWriter.Println("failed to delete Job with wrong input")
		e3 := f.DeleteJob("", namespace)
		Expect(e3).To(HaveOccurred())
		e3 = f.DeleteJob(jbName, "")
		Expect(e3).To(HaveOccurred())

		// failed to get Job with wrong input
		GinkgoWriter.Println("failed to get Job with wrong input")
		getjb4, e4 := f.GetJob("", namespace)
		Expect(getjb4).To(BeNil())
		Expect(e4).To(HaveOccurred())
		getjb4, e4 = f.GetJob(jbName, "")
		Expect(getjb4).To(BeNil())
		Expect(e4).To(HaveOccurred())

		// failed to get Job pod list with wrong input
		GinkgoWriter.Println("failed to get Job pod list with wrong input")
		podList5, e5 := f.GetJobPodList(jbNil)
		Expect(podList5).To(BeNil())
		Expect(e5).To(HaveOccurred())

	})
})
