// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework_test

import (
	"context"
	"fmt"
	"sync"
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

			Name: node.Name,
		},
	}
	key := client.ObjectKeyFromObject(fake)
	existing := &corev1.Node{}
	e := f.GetResource(key, existing)
	if e == nil && existing.DeletionTimestamp == nil {
		return fmt.Errorf("failed to create , a same node %v exists", node.Name)
	} else {
		t := func() bool {
			existing := &corev1.Node{}
			e := f.GetResource(key, existing)
			b := api_errors.IsNotFound(e)
			if !b {
				f.Log("waiting for a same node %v/%v to finish deleting \n", node.Name)
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
	var nodeName, nodeName2 string
	var label, label2 map[string]string
	var wg sync.WaitGroup
	BeforeEach(func() {
		f = fakeFramework()
		nodeName = "spider-worker"
		nodeName2 = "spider-worker2"
		label = map[string]string{
			"kubernetes.io/hostname": "spider-worker",
		}
		label2 = map[string]string{
			"kubernetes.io/hostname": "spider-worker2",
		}
	})

	It("operate nodelist", func() {

		wg.Add(1)
		go func() {
			GinkgoRecover()
			// create node yaml
			node1 := generateExampleNodeYaml(nodeName, label)
			Expect(node1).NotTo(BeNil())
			GinkgoWriter.Printf("finish creating node %v \n", nodeName)
			GinkgoWriter.Printf("node1: %v\n", node1)

			// create node
			e1 := createNode(f, node1)
			Expect(e1).NotTo(HaveOccurred())
			wg.Done()
		}()

		// wait cluster node ready
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		bll, errcheck := f.WaitClusterNodeReady(ctx)
		Expect(bll).NotTo(BeFalse())
		Expect(errcheck).NotTo(HaveOccurred())

		wg.Wait()

		// get node list
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

		// check node status
		bl := f.CheckNodeStatus(&nodeList.Items[0], true)
		Expect(bl).NotTo(BeFalse())

		// status is set to false in order to use node not ready
		wg.Add(1)
		go func() {
			GinkgoRecover()
			// create node yaml
			node2 := generateExampleNodeYaml(nodeName2, label2)
			Expect(node2).NotTo(BeNil())
			node2.Status.Conditions = []corev1.NodeCondition{
				{
					Status: corev1.ConditionFalse,
				},
			}
			GinkgoWriter.Printf("finish creating node %v \n", nodeName2)
			GinkgoWriter.Printf("node2: %v\n", node2)

			// create node
			e1 := createNode(f, node2)
			Expect(e1).NotTo(HaveOccurred())
			wg.Done()
		}()

		// wait cluster node ready
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel2()
		bll2, errcheck2 := f.WaitClusterNodeReady(ctx2)
		Expect(bll2).NotTo(BeFalse())
		Expect(errcheck2).NotTo(HaveOccurred())
		wg.Wait()

		// check node status
		bl2 := f.CheckNodeStatus(&nodeList.Items[0], true)
		Expect(bl2).To(BeTrue())

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
		bll := f.CheckNodeStatus(node3, false)
		Expect(bll).NotTo(BeTrue())

		// unit test WaitClusterNodeReady
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		bll, errcheck := f.WaitClusterNodeReady(ctx)
		Expect(bll).NotTo(BeFalse())
		Expect(errcheck).NotTo(HaveOccurred())
	})
})
