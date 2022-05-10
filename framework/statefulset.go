// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"fmt"
	"github.com/spidernet-io/e2eframework/tools"
	appsv1 "k8s.io/api/apps/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func (f *Framework) CreateStatefulSet(sts *appsv1.StatefulSet, opts ...client.CreateOption) error {
	// try to wait for finish last deleting
	fake := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sts.ObjectMeta.Namespace,
			Name:      sts.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &appsv1.StatefulSet{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same statefulSet %v/%v exists", sts.ObjectMeta.Namespace, sts.ObjectMeta.Name)
	} else {
		t := func() bool {
			existing := &appsv1.StatefulSet{}
			e := f.GetResource(key, existing)
			b := api_errors.IsNotFound(e)
			if !b {
				f.t.Logf("waiting for a same statefulSet %v/%v to finish deleting \n", sts.ObjectMeta.Namespace, sts.ObjectMeta.Name)
				return false
			}
			return true
		}
		if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
			return fmt.Errorf("time out to wait a deleting statefulset")
		}
	}

	return f.CreateResource(sts, opts...)
}

func (f *Framework) DeleteStatefulSet(name, namespace string, opts ...client.DeleteOption) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return f.DeleteResource(sts, opts...)
}

func (f *Framework) GetStatefulSet(name, namespace string) (*appsv1.StatefulSet, error) {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	key := client.ObjectKeyFromObject(sts)
	existing := &appsv1.StatefulSet{}
	e := f.GetResource(key, existing)
	if e != nil {
		return nil, e
	}
	return existing, e
}
