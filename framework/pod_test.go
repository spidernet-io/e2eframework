// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"time"

	"k8s.io/utils/pointer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	e2e "github.com/spidernet-io/e2eframework/framework"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateExamplePodYaml(podName, namespace string, label map[string]string, phase corev1.PodPhase) *corev1.Pod {
	Expect(podName).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      podName,
			Labels:    label,
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
		Status: corev1.PodStatus{
			Phase: phase,
		},
	}
}

var _ = Describe("test pod", Label("pod"), func() {
	var f *e2e.Framework
	var podName, namespace string
	var label map[string]string
	BeforeEach(func() {
		f = fakeFramework()

		podName = "testpod"
		namespace = "default"
		label = map[string]string{
			"app": "testpod",
		}
	})

	It("operate pod", func() {

		go func() {
			defer GinkgoRecover()
			// notice: WaitPodStarted use watch , but for the fake clientset,
			// the watch have started before the pod ready, or else the watch will miss the event
			// so we create the pod after WaitPodStarted
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			// generate pod yaml
			pod := generateExamplePodYaml(podName, namespace, label, "")

			// create pod
			e := f.CreatePod(pod)
			Expect(e).NotTo(HaveOccurred())
			GinkgoWriter.Printf("finish creating pod %v/%v \n", namespace, podName)
		}()

		// wait pod started
		ctx1, cancel1 := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel1()
		pod, e1 := f.WaitPodStarted(podName, namespace, ctx1)
		Expect(e1).NotTo(HaveOccurred())
		Expect(pod).NotTo(BeNil())

		// update pod status
		GinkgoWriter.Printf("update pod %v/%v status\n", namespace, podName)
		pod.Status.Phase = corev1.PodRunning
		Expect(f.UpdateResourceStatus(pod)).To(Succeed(), "failed to update pod %v/%v status\n", namespace, podName)

		// get pod
		getPod, e2 := f.GetPod(podName, namespace)
		Expect(e2).NotTo(HaveOccurred())
		Expect(getPod.Status.Phase).To(Equal(corev1.PodRunning), "failed to update pod %v/%v status\n", namespace, podName)
		GinkgoWriter.Printf("get pod: %+v \n", getPod)

		// get pod list
		podList, e3 := f.GetPodList(&client.ListOptions{Namespace: namespace})
		Expect(e3).NotTo(HaveOccurred())
		GinkgoWriter.Printf("len of pods: %v", len(podList.Items))

		// delete pod until finish
		ctx5, cancel5 := context.WithTimeout(context.Background(), time.Minute)
		defer cancel5()
		opts5 := &client.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}
		e4 := f.DeletePodUntilFinish(podName, namespace, ctx5, opts5)
		Expect(e4).NotTo(HaveOccurred())

		// the following are cases for testing pod list
		// generate pod yaml
		pod1 := generateExamplePodYaml(podName, namespace, label, "Running")

		// create pod
		e5 := f.CreatePod(pod1)
		Expect(e5).NotTo(HaveOccurred())
		GinkgoWriter.Printf("finish creating pod %v/%v \n", namespace, podName)

		// get pod list by label
		podList1, e6 := f.GetPodListByLabel(label)
		Expect(podList1).NotTo(BeNil())
		Expect(e6).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get pod list : %+v \n", podList1)

		// check pod list running
		ok1 := f.CheckPodListRunning(podList1)
		Expect(ok1).To(BeTrue())

		// wait pod list running
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel2()
		e7 := f.WaitPodListRunning(label, 1, ctx2)
		Expect(e7).NotTo(HaveOccurred())

		// delete pod list
		opts8 := &client.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}
		e8 := f.DeletePodList(podList1, opts8)
		Expect(e8).NotTo(HaveOccurred())

		// counter cases for testing pod list
		// generate pod yaml
		pod2 := generateExamplePodYaml(podName, namespace, label, "")

		// create pod
		e9 := f.CreatePod(pod2)
		Expect(e9).NotTo(HaveOccurred())
		GinkgoWriter.Printf("finish creating pod %v/%v \n", namespace, podName)

		// get pod list by label
		podList2, e10 := f.GetPodListByLabel(label)
		Expect(podList2).NotTo(BeNil())
		Expect(e10).NotTo(HaveOccurred())
		GinkgoWriter.Printf("get pod list : %+v \n", podList2)

		// check pod list running
		ok2 := f.CheckPodListRunning(podList2)
		Expect(ok2).To(BeFalse())

		// wait pod list running
		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel3()
		e11 := f.WaitPodListRunning(label, 1, ctx3)
		Expect(e11).To(HaveOccurred())
		e12 := f.WaitPodListRunning(label, 2, ctx3)
		Expect(e12).To(HaveOccurred())

		// delete podList repeatedly
		ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel4()
		opts4 := &client.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}
		e13 := f.DeletePodListRepeatedly(label, time.Second*2, ctx4, opts4)
		Expect(e13).To(BeNil())

		// wait components running
		ctx6, cancel6 := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel6()
		e14 := f.WaitNamespacePodRunning("kube-system", ctx6)
		Expect(e14).To(BeNil())
	})

	It("counter example with wrong input", func() {

		// failed wait pod ready with wrong input name/namespace to be empty
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		getPod1, err1 := f.WaitPodStarted("", namespace, ctx1)
		Expect(getPod1).To(BeNil())
		Expect(err1).Should(MatchError(e2e.ErrWrongInput))
		getPod2, err2 := f.WaitPodStarted(podName, "", ctx1)
		Expect(getPod2).To(BeNil())
		Expect(err2).Should(MatchError(e2e.ErrWrongInput))

		// UT cover get pod name/namespace input to be empty
		getPod3, err3 := f.GetPod("", namespace)
		Expect(getPod3).To(BeNil())
		Expect(err3).Should(MatchError(e2e.ErrWrongInput))
		getPod3, err3 = f.GetPod(podName, "")
		Expect(getPod3).To(BeNil())
		Expect(err3).Should(MatchError(e2e.ErrWrongInput))

		// UT cover delete pod name/namespace input to be empty
		err4 := f.DeletePod("", namespace)
		Expect(err4).Should(MatchError(e2e.ErrWrongInput))
		err4 = f.DeletePod(podName, "")
		Expect(err4).Should(MatchError(e2e.ErrWrongInput))

		// UT cover get pod list by label, input to be nil
		pods, err5 := f.GetPodListByLabel(nil)
		Expect(pods).To(BeNil())
		Expect(err5).To(MatchError(e2e.ErrWrongInput))

		// UT cover check pod list running, input to be nil
		ok1 := f.CheckPodListRunning(nil)
		Expect(ok1).To(BeFalse())

		// UT cover delete pod list, input to be nil
		opts6 := &client.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}
		err6 := f.DeletePodList(nil, opts6)
		Expect(err6).To(MatchError(e2e.ErrWrongInput))

		// UT wait pod list running, input to be nil
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel2()
		err7 := f.WaitPodListRunning(nil, 1, ctx2)
		Expect(err7).To(MatchError(e2e.ErrWrongInput))
		err8 := f.WaitPodListRunning(label, 0, ctx2)
		Expect(err8).To(MatchError(e2e.ErrWrongInput))

		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel3()
		opts3 := &client.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}
		err9 := f.DeletePodListRepeatedly(nil, time.Second*2, ctx3, opts3)
		Expect(err9).To(MatchError(e2e.ErrWrongInput))

		// delete pod until finish, with invalid input
		ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second)
		defer cancel4()
		opts4 := &client.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}
		err10 := f.DeletePodUntilFinish("", namespace, ctx4, opts4)
		Expect(err10).To(HaveOccurred())
		err11 := f.DeletePodUntilFinish(podName, "", ctx4, opts4)
		Expect(err11).To(HaveOccurred())

		// wait components running
		ctx6, cancel6 := context.WithTimeout(context.Background(), 0)
		defer cancel6()
		e14 := f.WaitNamespacePodRunning("", ctx6)
		Expect(e14).To(HaveOccurred())
		Expect(e14).Should(MatchError(e2e.ErrTimeOut))
	})
})
