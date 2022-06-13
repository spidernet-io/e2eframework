// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	e2e "github.com/spidernet-io/e2eframework/framework"
	"github.com/spidernet-io/e2eframework/tools"
	corev1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateExampleNodeYaml(nodeName string, label map[string]string) *corev1.Node {
	Expect(nodeName).NotTo(BeEmpty())
	Expect(label).NotTo(BeEmpty())

	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: label,
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
}

// in the unit test, no real node exists, so create a fake node
func createNode(f *e2e.Framework, node *corev1.Node, opts ...client.CreateOption) error {

	fake := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{

			Name: node.ObjectMeta.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &corev1.Node{}
	e := f.GetResource(key, existing)
	if e == nil && existing.ObjectMeta.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same node %v exists", node.ObjectMeta.Name)
	} else {
		t := func() bool {
			existing := &corev1.Node{}
			e := f.GetResource(key, existing)
			b := api_errors.IsNotFound(e)
			if !b {
				f.Log("waiting for a same node %v/%v to finish deleting \n", node.ObjectMeta.Name)
				return false
			}
			return true
		}
		if !tools.Eventually(t, f.Config.ResourceDeleteTimeout, time.Second) {
			return e2e.ErrTimeOut
		}
	}
	return f.CreateResource(node, opts...)
}

var _ = Describe("test node list ", Label("node"), func() {
	var f *e2e.Framework
	var nodeName string
	var label map[string]string

	BeforeEach(func() {
		f = fakeFramework()
		nodeName = "spider-worker"
		label = map[string]string{
			"kubernetes.io/hostname": "spider-worker",
		}
	})

	It("operate nodelist", func() {

		// create node yaml
		node1 := generateExampleNodeYaml(nodeName, label)
		Expect(node1).NotTo(BeNil())
		GinkgoWriter.Printf("finish creating node %v \n", nodeName)
		GinkgoWriter.Printf("node1: %v\n", node1)

		// create node
		e1 := createNode(f, node1)
		Expect(e1).NotTo(HaveOccurred())

		// getpodlist
		opts := []client.ListOption{
			client.MatchingLabels(map[string]string{
				"kubernetes.io/hostname": "spider-worker",
			}),
		}
		GinkgoWriter.Printf("opts: %v", opts)
		nodeList, e1 := f.GetNodeList(opts...)
		Expect(e1).NotTo(HaveOccurred())
		GinkgoWriter.Printf("nodeList: %v", nodeList)
		GinkgoWriter.Printf("corev1.NodeReady: %v", corev1.NodeReady)
		bl := f.CheckNodeStatus(&nodeList.Items[0], true)
		Expect(bl).NotTo(BeFalse())
	})

	It("counter example with wrong input", func() {
		// check node ready with null nodename
		ctx := context.TODO()
		node3 := &corev1.Node{}
		nodeName := ""
		err2 := f.KClient.Get(ctx, types.NamespacedName{Name: nodeName}, node3)
		GinkgoWriter.Printf("err2: %v\n", err2)
		bl := f.CheckNodeStatus(node3, true)
		Expect(bl).NotTo(BeTrue())

		// unit test IsClusterNodeReady
		bll, errcheck := f.IsClusterNodeReady()
		Expect(bll).NotTo(BeFalse())
		Expect(errcheck).NotTo(HaveOccurred())
	})
})
