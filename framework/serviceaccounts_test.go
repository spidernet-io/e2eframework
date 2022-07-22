// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"fmt"
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

var _ = Describe("ServiceAccounts", Label("ServiceAccounts"), func() {
	var (
		f                            *e2e.Framework
		saName, namespace, errSaName string
	)

	BeforeEach(func() {
		f = fakeFramework()
		saName = "test-serviceaccount"
		namespace = "test-namesapce"

		errSaName = "test-errSaName"

		// generate example serviceaccount obj
		saObj := generateExampleServiceAccountObj(saName, namespace)
		Expect(saObj).NotTo(BeNil())
		GinkgoWriter.Printf("saObj=%v/%v\n", saObj.Namespace, saObj.Name)

		// create service account
		Expect(createServiceAccount(f, saObj)).To(Succeed(), "failed to create serviceAccount %v%v\n", namespace, saName)
	})

	It("check service accounts ready", func() {
		// check service accounts ready
		Expect(f.CheckServiceAccountReady(saName, namespace, time.Second*10)).To(Succeed(), "timeout to wait service accounts %v ready\n", saName)

		// check service accounts ready with counter example

		err := f.CheckServiceAccountReady(errSaName, namespace, time.Second)
		Expect(err).Should(MatchError(e2e.ErrTimeOut))
	})

	// counter example with wrong input
	It("counter example with wrong input", func() {
		err := f.CheckServiceAccountReady("", namespace, time.Second*10)
		Expect(err).Should(MatchError(e2e.ErrWrongInput))

		err = f.CheckServiceAccountReady(saName, "", time.Second*10)
		Expect(err).Should(MatchError(e2e.ErrWrongInput))
	})
})

func createServiceAccount(f *e2e.Framework, serviceaccount *corev1.ServiceAccount, opts ...client.CreateOption) error {
	// try to wait for finish last deleting
	fake := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: serviceaccount.ObjectMeta.Namespace,
			Name:      serviceaccount.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &corev1.Pod{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same serviceaccount %v/%v exists", serviceaccount.ObjectMeta.Namespace, serviceaccount.ObjectMeta.Name)
	} else {
		t := func() bool {
			existing := &corev1.ServiceAccount{}
			e := f.GetResource(key, existing)
			b := api_errors.IsNotFound(e)
			if !b {
				f.Log("waiting for a same serviceaccount %v/%v to finish deleting \n", serviceaccount.ObjectMeta.Namespace, serviceaccount.ObjectMeta.Name)
				return false
			}
			return true
		}
		if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
			return e2e.ErrTimeOut
		}
	}
	return f.CreateResource(serviceaccount, opts...)
}

func generateExampleServiceAccountObj(saName, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: namespace,
		},
	}
}
