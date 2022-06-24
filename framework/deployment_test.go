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
		//the fake clientset will not schedule deployment replicaset,so mock the number
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: readyReplica,
			Replicas:      replica,
		},
	}
}

var _ = Describe("test deployment", Label("deployment"), func() {
	var f *e2e.Framework
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
	})

	It("operate deployment", func() {
		dpmName := "testDpm"
		namespace := "default"
		wrongDeployName := "xxx"
		replica := int32(3)
		readyReplica := int32(3)
		scaleReplicas := int32(2)
		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			// notice: WaitDeploymentReady use watch , but for the fake clientset,
			// the watch have started before the deployment ready, or else the watch will miss the event
			// so we create the deployment after WaitDeploymentReady
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			dpm := GenerateExampleDeploymentYaml(dpmName, namespace, replica, readyReplica)

			err1 := f.CreateDeployment(dpm)
			Expect(err1).NotTo(HaveOccurred())
			GinkgoWriter.Printf("finish creating deployment \n")

			wg.Done()
		}()

		// check wait deployment ready and check ip assign
		podList, errip := f.WaitDeploymentReadyAndCheckIP(dpmName, namespace, time.Second*30)
		Expect(errip).NotTo(HaveOccurred())
		Expect(podList).NotTo(BeNil())

		wg.Wait()

		// get deployment
		getDpm1, err3 := f.GetDeployment(dpmName, namespace)
		Expect(err3).NotTo(HaveOccurred())
		Expect(getDpm1).NotTo(BeNil())

		// check pod ipv4 ipv6
		err4 := f.CheckPodListIpReady(podList)
		Expect(err4).NotTo(HaveOccurred())

		// scale deployment
		GinkgoWriter.Println("scale deployment")
		getDpm2, err5 := f.ScaleDeployment(getDpm1, scaleReplicas)
		Expect(getDpm2).NotTo(BeNil())
		Expect(err5).NotTo(HaveOccurred())

		// create a deployment with a same name
		GinkgoWriter.Println("create a deployment with a same name")
		err6 := f.CreateDeployment(getDpm1)
		Expect(err6).To(HaveOccurred())
		GinkgoWriter.Printf("failed creating a deployment with a same name: %v\n", dpmName)

		// delete deployment util finish
		GinkgoWriter.Printf("delete deployment %v/%v util finish\n", namespace, dpmName)
		err7 := f.DeleteDeploymentUntilFinish(dpmName, namespace, time.Minute)
		Expect(err7).NotTo(HaveOccurred())
		GinkgoWriter.Printf("%v/%v deleted successfully\n", namespace, dpmName)

		// delete deployment util finish, with wrong deploymentName
		GinkgoWriter.Printf("delete wrong deployment %v/%v\n", namespace, wrongDeployName)
		err8 := f.DeleteDeploymentUntilFinish(wrongDeployName, namespace, time.Second)
		Expect(err8).To(HaveOccurred())
	})

	It("counter example with wrong input", func() {
		dpmName := "testDpm"
		namespace := "default"
		replica := int32(3)
		readyReplica := int32(3)
		scaleReplicas := int32(2)
		var dpmNil *appsv1.Deployment = nil
		dpm := GenerateExampleDeploymentYaml(dpmName, namespace, replica, readyReplica)

		// failed wait deployment ready with wrong input name/namespace to be empty
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		getdpm1, err1 := f.WaitDeploymentReady("", namespace, ctx1)
		Expect(getdpm1).To(BeNil())
		Expect(err1).Should(MatchError(e2e.ErrWrongInput))
		getdpm2, err2 := f.WaitDeploymentReady(dpmName, "", ctx1)
		Expect(getdpm2).To(BeNil())
		Expect(err2).Should(MatchError(e2e.ErrWrongInput))

		// UT cover get deployment name/namespace input to be empty
		getdpm3, err3 := f.GetDeployment("", namespace)
		Expect(getdpm3).To(BeNil())
		Expect(err3).Should(MatchError(e2e.ErrWrongInput))
		getdpm3, err3 = f.GetDeployment(dpmName, "")
		Expect(getdpm3).To(BeNil())
		Expect(err3).Should(MatchError(e2e.ErrWrongInput))

		// UT cover get deployment pod list input to be empty
		getdpm4, err4 := f.GetDeploymentPodList(dpmNil)
		Expect(getdpm4).To(BeNil())
		Expect(err4).Should(MatchError(e2e.ErrWrongInput))

		// UT cover scale deployment input to be empty
		getdpm5, err5 := f.ScaleDeployment(dpmNil, scaleReplicas)
		Expect(getdpm5).To(BeNil())
		Expect(err5).Should(MatchError(e2e.ErrWrongInput))

		// UT cover delete deployment name/namespace input to be empty
		err6 := f.DeleteDeployment("", namespace)
		Expect(err6).Should(MatchError(e2e.ErrWrongInput))
		err6 = f.DeleteDeployment(dpmName, "")
		Expect(err6).Should(MatchError(e2e.ErrWrongInput))

		// UT cover wait delete until complete with wrong input
		err7 := f.WaitPodListDeleted("", dpm.Spec.Selector.MatchLabels, ctx1)
		Expect(err7).Should(MatchError(e2e.ErrWrongInput))
		err7 = f.WaitPodListDeleted(namespace, nil, ctx1)
		Expect(err7).Should(MatchError(e2e.ErrWrongInput))

		// UT cover create deployment util ready
		deploy8, err8 := f.CreateDeploymentUntilReady(nil, time.Second*30)
		Expect(err8).Should(MatchError(e2e.ErrWrongInput))
		Expect(deploy8).To(BeNil())

		// UT cover delete deployment util finish
		err9 := f.DeleteDeploymentUntilFinish("", namespace, time.Second*30)
		Expect(err9).Should(MatchError(e2e.ErrWrongInput))
		err9 = f.DeleteDeploymentUntilFinish(dpmName, "", time.Second*30)
		Expect(err9).Should(MatchError(e2e.ErrWrongInput))

		// UT cover wait deployment ready and check ip assign ok
		podList, errip := f.WaitDeploymentReadyAndCheckIP("", "", time.Second*10)
		Expect(errip).Should(MatchError(e2e.ErrWrongInput))
		Expect(podList).To(BeNil())
	})
})
