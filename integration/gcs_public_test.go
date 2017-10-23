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
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"github.com/cloudfoundry/bosh-gcscli/client"
	"github.com/cloudfoundry/bosh-gcscli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCS Public Bucket", func() {
	Context("with read-only configuration", func() {
		var (
			ctx AssertContext
			cfg *config.GCSCli
		)

		BeforeEach(func() {
			var err error
			cfg, err = getPublicConfig()
			Expect(err).NotTo(HaveOccurred())

			ctx = NewAssertContext(AsDefaultCredentials)
			ctx.AddConfig(cfg)
		})
		AfterEach(func() {
			ctx.Cleanup()
		})

		Describe("with a public file", func() {
			BeforeEach(func() {
				Expect(ctx.Config.CredentialsSource).ToNot(Equal(config.NoneCredentialsSource),
					"Cannot use 'none' credentials to setup")

				session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
					"put", ctx.ContentFile, ctx.GCSFileName)
				Expect(err).ToNot(HaveOccurred())
				Expect(session.ExitCode()).To(BeZero())

				_, rwClient, err := client.NewSDK(*ctx.Config)
				Expect(err).ToNot(HaveOccurred())
				bucket := rwClient.Bucket(ctx.Config.BucketName)
				obj := bucket.Object(ctx.GCSFileName)
				err = obj.ACL().Set(context.Background(),
					storage.AllUsers, storage.RoleReader)
				Expect(err).ToNot(HaveOccurred())
			})

			It("can get", func() {
				roctx := ctx.Clone(AsReadOnlyCredentials)
				defer roctx.Cleanup()

				tmpLocalFile, err := ioutil.TempFile("", "gcscli-download")
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = os.Remove(tmpLocalFile.Name()) }()
				err = tmpLocalFile.Close()
				Expect(err).ToNot(HaveOccurred())

				session, err := RunGCSCLI(gcsCLIPath, roctx.ConfigPath,
					"get", ctx.GCSFileName, tmpLocalFile.Name())
				Expect(err).ToNot(HaveOccurred())
				Expect(session.ExitCode()).To(BeZero(),
					fmt.Sprintf("unexpected '%s'", session.Err.Contents()))

				gottenBytes, err := ioutil.ReadFile(tmpLocalFile.Name())
				Expect(err).ToNot(HaveOccurred())
				Expect(string(gottenBytes)).To(Equal(ctx.ExpectedString))
			})

			AfterEach(func() {
				session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
					"delete", ctx.GCSFileName)
				Expect(err).ToNot(HaveOccurred())
				Expect(session.ExitCode()).To(BeZero())
			})
		})

		It("fails to get a missing file", func() {
			roctx := ctx.Clone(AsReadOnlyCredentials)
			defer roctx.Cleanup()

			session, err := RunGCSCLI(gcsCLIPath, roctx.ConfigPath,
				"get", ctx.GCSFileName, "/dev/null")
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode()).ToNot(BeZero())
			Expect(session.Err.Contents()).To(ContainSubstring("object doesn't exist"))
		})

		It("fails to put", func() {
			Expect(ctx.Config.CredentialsSource).ToNot(Equal(config.NoneCredentialsSource),
				"Cannot use 'none' credentials to setup")

			roctx := ctx.Clone(AsReadOnlyCredentials)
			defer roctx.Cleanup()

			session, err := RunGCSCLI(gcsCLIPath, roctx.ConfigPath,
				"put", roctx.ContentFile, roctx.GCSFileName)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode()).ToNot(BeZero())
			Expect(session.Err.Contents()).To(ContainSubstring(client.ErrInvalidROWriteOperation.Error()))
		})

		It("fails to delete", func() {
			Expect(ctx.Config.CredentialsSource).ToNot(Equal(config.NoneCredentialsSource),
				"Cannot use 'none' credentials to setup")

			roctx := ctx.Clone(AsReadOnlyCredentials)
			defer roctx.Cleanup()

			session, err := RunGCSCLI(gcsCLIPath, roctx.ConfigPath,
				"delete", roctx.GCSFileName)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode()).ToNot(BeZero())
			Expect(session.Err.Contents()).To(ContainSubstring(client.ErrInvalidROWriteOperation.Error()))
		})
	})
})
