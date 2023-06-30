// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	"github.com/spidernet-io/e2eframework/tools"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateExampleEventYaml(eventKind, objName, objnamespace, message string) *corev1.Event {
	Expect(eventKind).NotTo(BeEmpty())
	Expect(objName).NotTo(BeEmpty())
	Expect(objnamespace).NotTo(BeEmpty())
	Expect(message).NotTo(BeEmpty())
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: objnamespace,
			Name:      objName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: eventKind,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      eventKind,
			Namespace: objnamespace,
			Name:      objName,
		},
		Message: message,
	}
}

// WaitExceptEventOccurred unit test, because there is no real event, so create a fake event through the CreateEvent
func CreateEvent(f *e2e.Framework, event *corev1.Event, opts ...client.CreateOption) error {
	fake := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: event.ObjectMeta.Namespace,
			Name:      event.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &corev1.Event{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same event %v exists", event.ObjectMeta.Name)
	}
	t := func() bool {
		existing := &corev1.Event{}
		e := f.GetResource(key, existing)
		b := api_errors.IsNotFound(e)
		if !b {
			f.Log("waiting for a same event %v/%v to finish deleting \n", event.ObjectMeta.Name, event.ObjectMeta.Namespace)
			return false
		}
		return true
	}
	if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
		return e2e.ErrTimeOut
	}
	return f.CreateResource(event, opts...)
}

var _ = Describe("test events", Label("events"), func() {
	var f *e2e.Framework
	var eventKind, podName, namespace, message, nonExistingMessage string
	var label map[string]string
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
		eventKind = "Pod"
		podName = "test-events"
		namespace = "default"
		message = "test"
		nonExistingMessage = "non-existing"
		label = map[string]string{
			"app": "test-events",
		}
	})

	It("operate events", func() {
		// generate pod yaml
		pod := generateExamplePodYaml(podName, namespace, label, "")

		// Constructing events by creating pods
		e := f.CreatePod(pod)
		Expect(e).NotTo(HaveOccurred())
		GinkgoWriter.Printf("finish creating pod %v/%v \n", namespace, podName)

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			// notice: WaitExceptEventOccurred use watch , but for the fake clientset,
			// the watch have started before the pod ready, or else the watch will miss the event
			// so we create the pod after WaitExceptEventOccurred
			// in the real environment, this issue does not exist
			time.Sleep(2 * time.Second)
			// create event
			event := generateExampleEventYaml(eventKind, podName, namespace, message)
			e = CreateEvent(f, event)
			Expect(e).NotTo(HaveOccurred())
			wg.Done()
		}()
		// check event
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel1()
		e1 := f.WaitExceptEventOccurred(ctx1, eventKind, podName, namespace, message)
		Expect(e1).NotTo(HaveOccurred())
		// Wait except event occurred time out
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel2()
		e2 := f.WaitExceptEventOccurred(ctx2, eventKind, podName, namespace, nonExistingMessage)
		Expect(e2).To(HaveOccurred())

		// get all events
		_, err := f.GetEvents(context.Background(), eventKind, podName, namespace)
		Expect(err).NotTo(HaveOccurred())

		// delete pod
		e3 := f.DeletePod(podName, namespace)
		Expect(e3).NotTo(HaveOccurred())
	})
	It("counter example with wrong input", func() {
		ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel1()
		e4 := f.WaitExceptEventOccurred(ctx1, eventKind, "", namespace, message)
		Expect(e4).To(HaveOccurred())
		Expect(e4).Should(MatchError(e2e.ErrWrongInput))

		e5 := f.WaitExceptEventOccurred(ctx1, eventKind, podName, "", message)
		Expect(e5).To(HaveOccurred())
		Expect(e5).Should(MatchError(e2e.ErrWrongInput))

		e6 := f.WaitExceptEventOccurred(ctx1, eventKind, podName, namespace, "")
		Expect(e6).To(HaveOccurred())
		Expect(e6).Should(MatchError(e2e.ErrWrongInput))

		e7 := f.WaitExceptEventOccurred(ctx1, "", podName, namespace, message)
		Expect(e7).To(HaveOccurred())
		Expect(e7).Should(MatchError(e2e.ErrWrongInput))
	})

})
