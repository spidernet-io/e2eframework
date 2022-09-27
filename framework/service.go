// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"github.com/spidernet-io/e2eframework/tools"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Framework) CreateService(service *corev1.Service, opts ...client.CreateOption) error {
	if service == nil {
		return ErrWrongInput
	}
	// try to wait for finish last deleting
	key := types.NamespacedName{
		Name:      service.ObjectMeta.Name,
		Namespace: service.ObjectMeta.Namespace,
	}
	existing := &corev1.Service{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same service %v/%v exists", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
	} else {
		t := func() bool {
			existing := &corev1.Pod{}
			e := f.GetResource(key, existing)
			b := api_errors.IsNotFound(e)
			if !b {
				f.Log("waiting for a same service %v/%v to finish deleting \n", service.ObjectMeta.Namespace, service.ObjectMeta.Name)
				return false
			}
			return true
		}
		if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
			return ErrTimeOut
		}
	}
	return f.CreateResource(service, opts...)
}

func (f *Framework) GetService(name, namespace string) (*corev1.Service, error) {
	if name == "" || namespace == "" {
		return nil, ErrWrongInput
	}
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	service := &corev1.Service{}
	err := f.GetResource(key, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (f *Framework) WaitServiceReady(name, namespace string, timeout time.Duration) error {
	if name == "" || namespace == "" {
		return ErrWrongInput
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return ErrTimeOut
		default:
			service, err := f.GetService(name, namespace)
			if err != nil {
				return err
			}
			if service != nil {
				return nil
			}
			time.Sleep(time.Second)
		}
	}
}
