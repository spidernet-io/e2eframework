// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/spidernet-io/e2eframework/tools"

	//appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const daemonNs = "ds.ObjectMeta.Namespace"
const daemonName = "ds.ObjectMeta.Name"

func (f *Framework) CreateJob(jb *batchv1.Job, opts ...client.CreateOption) error {

	// try to wait for finish last deleting
	fake := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: jb.ObjectMeta.Namespace,
			Name:      jb.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &batchv1.Job{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same Job %v/%v exist", daemonNs, daemonName)
	}
	t := func() bool {
		existing := &batchv1.Job{}
		e := f.GetResource(key, existing)
		b := api_errors.IsNotFound(e)
		if !b {
			f.t.Logf("waiting for a same Job %v/%v to finish deleting \n", daemonNs, daemonName)
			return false
		}
		return true
	}
	if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
		return ErrTimeOutWait
	}

	return f.CreateResource(jb, opts...)
}

func (f *Framework) DeleteJob(name, namespace string, opts ...client.DeleteOption) error {
	if name == "" || namespace == "" {
		return ErrWrongInput

	}

	jb := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	return f.DeleteResource(jb, opts...)
}

func (f *Framework) GetJob(name, namespace string) (*batchv1.Job, error) {
	if name == "" || namespace == "" {
		return nil, ErrWrongInput
	}

	jb := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	key := client.ObjectKeyFromObject(jb)
	existing := &batchv1.Job{}
	e := f.GetResource(key, existing)
	if e != nil {
		return nil, e
	}
	return existing, e
}

func (f *Framework) GetJobPodList(jb *batchv1.Job) (*corev1.PodList, error) {
	if jb == nil {
		return nil, ErrWrongInput
	}
	pods := &corev1.PodList{}
	ops := []client.ListOption{
		client.MatchingLabelsSelector{
			Selector: labels.SelectorFromSet(jb.Spec.Selector.MatchLabels),
		},
	}
	e := f.ListResource(pods, ops...)
	if e != nil {
		return nil, e
	}
	return pods, nil
}

func (f *Framework) WaitJobReady(name, namespace string, ctx context.Context) (*batchv1.Job, error) {
	if name == "" || namespace == "" {
		return nil, ErrWrongInput
	}

	l := &client.ListOptions{
		Namespace:     namespace,
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name),
	}
	watchInterface, err := f.KClient.Watch(ctx, &batchv1.JobList{}, l)

	if err != nil {
		return nil, ErrWatch
	}
	defer watchInterface.Stop()

	for {
		select {
		// if jb not exist , got no event
		case event, ok := <-watchInterface.ResultChan():
			if !ok {
				return nil, ErrWrongInput
			}
			f.t.Logf(" jb %v/%v %v event \n", namespace, name, event.Type)

			switch event.Type {
			case watch.Error:
				return nil, ErrEvent
			case watch.Deleted:
				return nil, ErrResDel
			default:
				jb, ok := event.Object.(*batchv1.Job)
				if !ok {
					return nil, ErrGetObj
				}

				if jb.Status.Active == 0 {
					break

				} else if jb.Status.Active == *(jb.Spec.Parallelism) {
					time.Sleep(time.Second * 20)
					return jb, nil
				}
			}
		case <-ctx.Done():
			return nil, ErrTimeOut
		}
	}
}

func (f *Framework) WaitJobFinished(jobName, namespace string, ctx context.Context) (*batchv1.Job, error) {
	for {
		select {
		default:
			job, err := f.GetJob(jobName, namespace)
			if err != nil {
				return nil, err
			}
			for _, c := range job.Status.Conditions {
				if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
					return job, nil
				}
			}
			time.Sleep(time.Second)
		case <-ctx.Done():
			return nil, ErrTimeOut

		}
	}
}
