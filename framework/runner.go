// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0
package framework

import (
	"context"
	"k8s.io/klog/v2"
	"k8s.io/utils/exec"
)

var Kubectl = "kubectl"

// Interface is an injectable interface for running  commands
type Interface interface {
	// KubectlCmd is use for execute kubectl command
	KubectlCmd(ctx context.Context, args ...string) ([]byte, error)
	// KindCmd is use for execute kind command
	KindCmd(ctx context.Context, args ...string) ([]byte, error)
	// VagrantCmd is use for execute vagrant command
	VagrantCmd(ctx context.Context, args ...string) ([]byte, error)
	// TODO: other cmd
}

// NewRunner return an instance of Interface
func NewRunner() Interface {
	return &runner{
		execer: exec.New(),
	}
}

type runner struct {
	execer exec.Interface
}

// KubectlCmd execute a kubectl command
func (r *runner) KubectlCmd(ctx context.Context, args ...string) ([]byte, error) {
	klog.Infof("run cmd: %s %v ", Kubectl, args)
	if ctx == nil {
		return r.execer.Command(Kubectl, args...).CombinedOutput()
	}
	return r.execer.CommandContext(ctx, Kubectl, args...).CombinedOutput()
}

// KindCmd execute a kind command
func (r *runner) KindCmd(ctx context.Context, args ...string) ([]byte, error) {
	return nil, nil
}

// VagrantCmd execute a vagrant command
func (r *runner) VagrantCmd(ctx context.Context, args ...string) ([]byte, error) {
	return nil, nil
}
