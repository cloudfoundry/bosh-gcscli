/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package integration

import (
	"github.com/cloudfoundry/bosh-gcscli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Integration", func() {
	Context("general (Default Applicaton Credentials) configuration", func() {
		var (
			ctx AssertContext
			cfg *config.GCSCli
		)
		BeforeEach(func() {
			ctx = NewAssertContext(AsDefaultCredentials)
			cfg = getMultiRegionConfig()
			cfg.EncryptionKey = encryptionKeyBytes

		})
		AfterEach(func() {
			ctx.Cleanup()
		})

		encryptedConfigs := getEncryptedConfigs()

		DescribeTable("Get with correct encryption_key works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertEncryptionWorks(gcsCLIPath, ctx)
			},
			encryptedConfigs...)

		DescribeTable("Get with wrong encryption_key should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertWrongKeyEncryptionFails(gcsCLIPath, ctx)
			},
			encryptedConfigs...)

		DescribeTable("Get with no encryption_key should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertNoKeyEncryptionFails(gcsCLIPath, ctx)
			},
			encryptedConfigs...)
	})
})
