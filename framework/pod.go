// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/spidernet-io/e2eframework/tools"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Framework) CreatePod(pod *corev1.Pod, opts ...client.CreateOption) error {
	// try to wait for finish last deleting
	fake := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pod.ObjectMeta.Namespace,
			Name:      pod.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &corev1.Pod{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same pod %v/%v exists", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
	} else {
		t := func() bool {
			existing := &corev1.Pod{}
			e := f.GetResource(key, existing)
			b := api_errors.IsNotFound(e)
			if !b {
				f.t.Logf("waiting for a same pod %v/%v to finish deleting \n", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
				return false
			}
			return true
		}
		if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
			return fmt.Errorf("time out to wait a deleting pod")
		}
	}
	return f.CreateResource(pod, opts...)
}

func (f *Framework) DeletePod(name, namespace string, opts ...client.DeleteOption) error {

	if name == "" || namespace == "" {
		return ErrWrongInput
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return f.DeleteResource(pod, opts...)
}

func (f *Framework) GetPod(name, namespace string) (*corev1.Pod, error) {

	if name == "" || namespace == "" {
		return nil, ErrWrongInput
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	key := client.ObjectKeyFromObject(pod)
	existing := &corev1.Pod{}
	e := f.GetResource(key, existing)
	if e != nil {
		return nil, e
	}
	return existing, e
}

func (f *Framework) GetPodList(opts ...client.ListOption) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	e := f.ListResource(pods, opts...)
	if e != nil {
		return nil, e
	}
	return pods, nil
}

func (f *Framework) WaitPodStarted(name, namespace string, ctx context.Context) (*corev1.Pod, error) {

	if name == "" || namespace == "" {
		return nil, ErrWrongInput
	}

	// refer to https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/client/watch_test.go
	l := &client.ListOptions{
		Namespace:     namespace,
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name),
	}
	watchInterface, err := f.KClient.Watch(ctx, &corev1.PodList{}, l)
	if err != nil {
		return nil, fmt.Errorf("failed to Watch: %v", err)
	}
	defer watchInterface.Stop()

	for {
		select {
		// if pod not exist , got no event
		case event, ok := <-watchInterface.ResultChan():
			if !ok {
				return nil, fmt.Errorf("channel is closed ")
			}
			f.t.Logf("pod %v/%v %v event \n", namespace, name, event.Type)
			// Added    EventType = "ADDED"
			// Modified EventType = "MODIFIED"
			// Deleted  EventType = "DELETED"
			// Bookmark EventType = "BOOKMARK"
			// Error    EventType = "ERROR"
			switch event.Type {
			case watch.Error:
				return nil, fmt.Errorf("received error event: %+v", event)
			case watch.Deleted:
				return nil, fmt.Errorf("resource is deleted")
			default:
				pod, ok := event.Object.(*corev1.Pod)
				// metaObject, ok := event.Object.(metav1.Object)
				if !ok {
					return nil, fmt.Errorf("failed to get metaObject")
				}
				f.t.Logf("pod %v/%v status=%+v\n", namespace, name, pod.Status.Phase)
				if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodUnknown {
					break
				} else {
					return pod, nil
				}
			}
		case <-ctx.Done():
			return nil, ErrTimeOut
		}
	}
}

func (f *Framework) WaitDeleteUntilComplete(namespace string, label map[string]string, ctx context.Context) error {
	// Query all pods corresponding to the label
	// Delete the resource until the query is empty

	if namespace == "" || label == nil {
		return ErrWrongInput
	}

	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(label),
	}
	for {
		select {
		case <-ctx.Done():
			return ErrTimeOut
		default:
			podlist, err := f.GetPodList(opts...)
			if err != nil {
				return err
			} else if len(podlist.Items) == 0 {
				return nil
			}
		}
	}
}

func (f *Framework) CheckPodListIpReady(podlist *corev1.PodList) error {

	for i := 0; i < len(podlist.Items); i++ {
		if podlist.Items[i].Status.PodIPs == nil {
			return fmt.Errorf("pod %v failed to assign ip", podlist.Items[i].Name)
		}
		f.t.Logf("pod %v ips: %+v \n", podlist.Items[i].Name, podlist.Items[i].Status.PodIPs)

		if !f.Info.IpV4Enabled || !tools.CheckPodIpv4IPReady(&podlist.Items[i]) {
			return fmt.Errorf("pod %v failed to get ipv4 ip", podlist.Items[i].Name)
		}
		f.t.Logf("succeeded to check pod %v ipv4 ip \n", podlist.Items[i].Name)

		if !f.Info.IpV6Enabled || !tools.CheckPodIpv6IPReady(&podlist.Items[i]) {
			return fmt.Errorf("pod %v failed to get ipv6 ip", podlist.Items[i].Name)
		}
		f.t.Logf("succeeded to check pod %v ipv6 ip \n", podlist.Items[i].Name)
	}
	return nil
}
