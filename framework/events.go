// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Framework) ListEvents(opts ...client.ListOption) (events *corev1.EventList, err error) {
	events = &corev1.EventList{}
	e := f.ListResource(events, opts...)
	if e != nil {
		return nil, e
	}
	return events, nil
}
