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

func GenerateExampleReplicaSetYaml(rsName, namespace string, replica, readyReplica int32) *appsv1.ReplicaSet {
	Expect(rsName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      rsName,
		},

		Spec: appsv1.ReplicaSetSpec{
			Replicas: pointer.Int32Ptr(replica),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": rsName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": rsName,
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
		//the fake clientset will not schedule replicaset,so mock the number
		Status: appsv1.ReplicaSetStatus{
			ReadyReplicas: readyReplica,
			//	Replicas:      replica,
		},
	}
}

var _ = Describe("test ReplicaSet", Label("ReplicaSet"), func() {
	var f *e2e.Framework
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate ReplicaSet", func() {
		rsName := "testrs"
		namespace := "rs-ns"
		replica := int32(3)
		readyReplica := int32(3)
		scaleReplicas := int32(2)
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			// notice: WaitPodStarted use watch , but for the fake clientset,
			// the watch have started before the pod ready, or else the watch will miss the event
			// so we create the pod after WaitPodStarted
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			rs := GenerateExampleReplicaSetYaml(rsName, namespace, replica, readyReplica)

			err1 := f.CreateReplicaSet(rs)
			Expect(err1).NotTo(HaveOccurred())
			GinkgoWriter.Printf("finish creating ReplicaSet \n")

			wg.Done()
		}()

		// wait ReplicaSet ready
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		rs, err2 := f.WaitReplicaSetReady(rsName, namespace, ctx1)
		Expect(err2).NotTo(HaveOccurred())
		Expect(rs).NotTo(BeNil())

		wg.Wait()

		// get ReplicaSet
		getrs1, err3 := f.GetReplicaSet(rsName, namespace)
		Expect(err3).NotTo(HaveOccurred())
		Expect(getrs1).NotTo(BeNil())

		// get ReplicaSet pod list
		GinkgoWriter.Println("get ReplicaSet pod list")
		podList, err4 := f.GetReplicaSetPodList(rs)
		Expect(podList).NotTo(BeNil())
		Expect(err4).NotTo(HaveOccurred())

		// scale ReplicaSet
		GinkgoWriter.Println("scale ReplicaSet")
		getrs2, err5 := f.ScaleReplicaSet(rs, scaleReplicas)
		Expect(getrs2).NotTo(BeNil())
		Expect(err5).NotTo(HaveOccurred())

		// create a ReplicaSet with a same name
		GinkgoWriter.Println("create a ReplicaSet with a same name")
		err6 := f.CreateReplicaSet(rs)
		Expect(err6).To(HaveOccurred())
		GinkgoWriter.Printf("failed creating a ReplicaSet with a same name: %v\n", rsName)

		// delete ReplicaSet
		GinkgoWriter.Printf("delete ReplicaSet %v \n", rsName)
		err7 := f.DeleteReplicaSet(rsName, namespace)
		Expect(err7).NotTo(HaveOccurred())
		GinkgoWriter.Printf("%v deleted successfully \n", rsName)
	})

	It("counter example with wrong input", func() {
		rsName := "testrs"
		namespace := "rs-ns"
		scaleReplicas := int32(2)
		var rsNil *appsv1.ReplicaSet = nil

		// failed wait ReplicaSet ready with wrong input name/namespace to be empty
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		getrs1, err1 := f.WaitReplicaSetReady("", namespace, ctx1)
		Expect(getrs1).To(BeNil())
		Expect(err1).To(HaveOccurred())
		getrs2, err2 := f.WaitReplicaSetReady(rsName, "", ctx1)
		Expect(getrs2).To(BeNil())
		Expect(err2).To(HaveOccurred())

		// UT cover get ReplicaSet name/namespace input to be empty
		getrs3, err3 := f.GetDeploymnet("", namespace)
		Expect(getrs3).To(BeNil())
		Expect(err3).To(HaveOccurred())
		getrs3, err3 = f.GetDeploymnet(rsName, "")
		Expect(getrs3).To(BeNil())
		Expect(err3).To(HaveOccurred())

		// UT cover get ReplicaSet pod list input to be empty
		getrs4, err4 := f.GetReplicaSetPodList(rsNil)
		Expect(getrs4).To(BeNil())
		Expect(err4).To(HaveOccurred())

		// UT cover scale ReplicaSet input to be empty
		getrs5, err5 := f.ScaleReplicaSet(rsNil, scaleReplicas)
		Expect(getrs5).To(BeNil())
		Expect(err5).To(HaveOccurred())

		// UT cover delete ReplicaSet name/namespace input to be empty
		err6 := f.DeleteReplicaSet("", namespace)
		Expect(err6).To(HaveOccurred())
		err6 = f.DeleteReplicaSet(rsName, "")
		Expect(err6).To(HaveOccurred())
	})
})
