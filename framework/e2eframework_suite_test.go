// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"testing"

	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spiderv2beta1 "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v2beta1"
	batchv1 "k8s.io/api/batch/v1"

	multus_v1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	e2e "github.com/spidernet-io/e2eframework/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensions_v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestE2eframework(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2eframework Suite")
}

// ========================

func fakeClientSet() client.WithWatch {
	scheme := runtime.NewScheme()
	err := batchv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = apiextensions_v1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = appsv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = apiextensions_v1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = multus_v1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = spiderv2beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&multus_v1.NetworkAttachmentDefinition{}).Build()
}

func fakeKubeConfig() *os.File {
	file, err := os.CreateTemp("", "unitest")
	Expect(err).NotTo(HaveOccurred())
	// /var/folders/zk/13j3111s12n04r0pcx_sky_w0000gn/T/unitest1183784803
	GinkgoWriter.Printf("fake a kubeconfig file %v", file.Name())
	return file
}

func fakeEnv(kubeconfigpath string) {
	os.Setenv(e2e.E2E_CLUSTER_NAME, "testcluster")
	os.Setenv(e2e.E2E_KUBECONFIG_PATH, kubeconfigpath)
	os.Setenv(e2e.E2E_IPV4_ENABLED, "true")
	os.Setenv(e2e.E2E_IPV6_ENABLED, "true")
	os.Setenv(e2e.E2E_MULTUS_CNI_ENABLED, "true")
	os.Setenv(e2e.E2E_SPIDERPOOL_IPAM_ENABLED, "true")
	os.Setenv(e2e.E2E_WHEREABOUT_IPAM_ENABLED, "false")
	os.Setenv(e2e.E2E_KIND_CLUSTER_NODE_LIST, "master,worker")
}

func clearEnv() {
	os.Setenv(e2e.E2E_CLUSTER_NAME, "")
	os.Setenv(e2e.E2E_KUBECONFIG_PATH, "")
	os.Setenv(e2e.E2E_IPV4_ENABLED, "")
	os.Setenv(e2e.E2E_IPV6_ENABLED, "")
	os.Setenv(e2e.E2E_MULTUS_CNI_ENABLED, "")
	os.Setenv(e2e.E2E_SPIDERPOOL_IPAM_ENABLED, "")
	os.Setenv(e2e.E2E_WHEREABOUT_IPAM_ENABLED, "")
	os.Setenv(e2e.E2E_KIND_CLUSTER_NODE_LIST, "")
}

// ========================

func fakeFramework() *e2e.Framework {
	fakeEnv("/tmp/nokubeconfigfile")

	fakeClient := fakeClientSet()
	Expect(fakeClient).NotTo(BeNil())

	f, e := e2e.NewFramework(GinkgoT(), nil, fakeClient)
	Expect(e).NotTo(HaveOccurred())

	return f
}
