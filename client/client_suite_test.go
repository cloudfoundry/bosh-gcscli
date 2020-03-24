// This file was adapted from: https://github.com/cloudfoundry/bosh-google-cpi-release
//
// Copyright (c) 2015-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// This project is licensed to you under the Apache License, Version 2.0 (the "License").
//
// You may not use this project except in compliance with the License.
//
// This project may include a number of subcomponents with separate copyright notices
// and license terms. Your use of these subcomponents is subject to the terms and
// conditions of the subcomponent's license, as noted in the LICENSE file.

package client_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoogleClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}
