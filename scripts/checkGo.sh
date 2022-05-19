#!/bin/bash

# Copyright 2022 Authors of spidernet-io
# SPDX-License-Identifier: Apache-2.0

SCRIPT_DIR=$( cd $( dirname "$0" ) && pwd )
PROJECT_DIR=$( cd ${SCRIPT_DIR}/.. && pwd )

cd ${PROJECT_DIR}

# expect not to use gomega in framework code
grep "github.com/onsi/gomega"  *  -RHn --colour \
  --include=*.go \
  --exclude=*_test.go --exclude-dir=vendor  2>/dev/null

(($?==0)) && exit 1

exit 0
