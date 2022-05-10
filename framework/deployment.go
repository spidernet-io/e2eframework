// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/spidernet-io/e2eframework/tools"
	appsv1 "k8s.io/api/apps/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Framework) CreateDeployment(dpm *appsv1.Deployment, opts ...client.CreateOption) error {
	// try to wait for finish last deleting
	fake := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dpm.ObjectMeta.Namespace,
			Name:      dpm.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &appsv1.Deployment{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same deployment %v/%v exists", dpm.ObjectMeta.Namespace, dpm.ObjectMeta.Name)
	}
	t := func() bool {
		existing := &appsv1.Deployment{}
		e := f.GetResource(key, existing)
		b := api_errors.IsNotFound(e)
		if !b {
			f.t.Logf("waiting for a same deployment %v/%v to finish deleting \n", dpm.ObjectMeta.Namespace, dpm.ObjectMeta.Name)
			return false
		}
		return true
	}
	if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
		return fmt.Errorf("time out to wait a deleting deployment")
	}
	return f.CreateResource(dpm, opts...)
}

func (f *Framework) DeleteDeployment(name, namespace string, opts ...client.DeleteOption) error {
	Expect(name).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())
	pod := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return f.DeleteResource(pod, opts...)
}

func (f *Framework) GetDeploymnet(name, namespace string) (*appsv1.Deployment, error) {
	Expect(name).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	dpm := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	key := client.ObjectKeyFromObject(dpm)
	existing := &appsv1.Deployment{}
	e := f.GetResource(key, existing)
	if e != nil {
		return nil, e
	}
	return existing, e
}

func (f *Framework) WaitDeploymentReady(name, namespace string, ctx context.Context) (*appsv1.Deployment, error) {
	Expect(name).NotTo(BeEmpty())
	Expect(namespace).NotTo(BeEmpty())

	l := &client.ListOptions{
		Namespace:     namespace,
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name),
	}
	watchInterface, err := f.KClient.Watch(ctx, &appsv1.DeploymentList{}, l)
	if err != nil {
		return nil, fmt.Errorf("failed to Watch: %v", err)
	}
	defer watchInterface.Stop()

	for {
		select {
		case event, ok := <-watchInterface.ResultChan():
			f.t.Logf("deployment %v/%v\n", event, ok)
			if !ok {
				return nil, fmt.Errorf("channel is closed ")
			}
			f.t.Logf("deployment %v/%v %v event \n", namespace, name, event.Type)
			switch event.Type {
			case watch.Error:
				return nil, fmt.Errorf("received error event: %+v", event)
			case watch.Deleted:
				return nil, fmt.Errorf("resource is deleted")
			default:
				dpm, ok := event.Object.(*appsv1.Deployment)
				if !ok {
					return nil, fmt.Errorf("failed to get metaObject")
				}
				f.t.Logf("deployment %v/%v readyReplicas=%+v\n", namespace, name, dpm.Status.ReadyReplicas)
				if dpm.Status.ReadyReplicas == *(dpm.Spec.Replicas) {
					return dpm, nil
				}
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("ctx timeout ")
		}
	}
}
