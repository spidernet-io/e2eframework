// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Framework) GetServiceAccount(saName, namespace string) (*corev1.ServiceAccount, error) {
	if saName == "" || namespace == "" {
		return nil, ErrWrongInput
	}

	key := client.ObjectKey{
		Namespace: namespace,
		Name:      saName,
	}
	existing := &corev1.ServiceAccount{}
	e := f.GetResource(key, existing)
	if e != nil {
		return nil, e
	}
	return existing, e
}

func (f *Framework) WaitServiceAccountReady(saName, namespace string, timeout time.Duration) error {
	if saName == "" || namespace == "" {
		return ErrWrongInput
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		default:
			as, err := f.GetServiceAccount(saName, namespace)
			b := api_errors.IsNotFound(err)
			if b {
				f.Log("service account: %s/%s not found", namespace, saName)
				time.Sleep(time.Second)
				continue
			}

			if err != nil {
				f.Log("failed to get service account, error %v ", err)
				time.Sleep(time.Second)
				continue
			}

			if as != nil {
				return nil
			}

		case <-ctx.Done():
			return fmt.Errorf("%w: failed to wait for service account to be ready", ErrTimeOut)
		}
	}
}
