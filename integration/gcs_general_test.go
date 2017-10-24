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
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/cloudfoundry/bosh-gcscli/client"
	"github.com/cloudfoundry/bosh-gcscli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

// randReadSeeker is a ReadSeeker which returns random content and
// non-nil error for every operation.
//
// crypto/rand is used to ensure any compression
// applied to the reader's output doesn't effect the work we intend to do.
type randReadSeeker struct {
	reader io.Reader
}

func newrandReadSeeker(maxSize int64) randReadSeeker {
	limited := io.LimitReader(rand.Reader, maxSize)
	return randReadSeeker{limited}
}

func (rrs *randReadSeeker) Read(p []byte) (n int, err error) {
	return rrs.reader.Read(p)
}

func (rrs *randReadSeeker) Seek(offset int64, whenc int) (n int64, err error) {
	return offset, nil
}

// badReadSeeker is a ReadSeeker which returns a non-nil error
// for every operation.
type badReadSeeker struct{}

var badReadSeekerErr = io.ErrUnexpectedEOF

func (brs *badReadSeeker) Read(p []byte) (n int, err error) {
	return 0, badReadSeekerErr
}

func (brs *badReadSeeker) Seek(offset int64, whenc int) (n int64, err error) {
	return 0, badReadSeekerErr
}

var _ = Describe("Integration", func() {
	Context("general (Default Applicaton Credentials) configuration", func() {
		var ctx AssertContext
		BeforeEach(func() {
			ctx = NewAssertContext(AsDefaultCredentials)
		})
		AfterEach(func() {
			ctx.Cleanup()
		})

		configurations := getBaseConfigs()

		DescribeTable("Blobstore lifecycle works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertLifecycleWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Delete silently ignores that the file doesn't exist",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)

				session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
					"delete", ctx.GCSFileName)
				Expect(err).ToNot(HaveOccurred())
				Expect(session.ExitCode()).To(BeZero())
			},
			configurations...)

		// Perform a large file put causing GCS to do a multi-part upload
		DescribeTable("Multipart Put works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				if os.Getenv(NoLongEnv) != "" {
					Skip(fmt.Sprintf(NoLongMsg, NoLongEnv))
				}

				const twoGB = 1024 * 1024 * 1024 * 2
				limited := newrandReadSeeker(twoGB)

				_, gcsClient, err := client.NewSDK(*ctx.Config)
				Expect(err).ToNot(HaveOccurred())
				blobstoreClient, err := client.New(context.Background(),
					gcsClient, ctx.Config)
				Expect(err).ToNot(HaveOccurred())

				err = blobstoreClient.Put(&limited, ctx.GCSFileName)
				Expect(err).ToNot(HaveOccurred())

				blobstoreClient.Delete(ctx.GCSFileName)
				Expect(err).ToNot(HaveOccurred())
			},
			configurations...)

		DescribeTable("Invalid Put should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)

				_, gcsClient, err := client.NewSDK(*ctx.Config)
				Expect(err).ToNot(HaveOccurred())
				blobstoreClient, err := client.New(context.Background(),
					gcsClient, ctx.Config)
				Expect(err).ToNot(HaveOccurred())

				err = blobstoreClient.Put(&badReadSeeker{}, ctx.GCSFileName)
				Expect(err).To(HaveOccurred())
			},
			configurations...)

		DescribeTable("Invalid Get should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)

				session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
					"get", ctx.GCSFileName, "/dev/null")
				Expect(err).ToNot(HaveOccurred())
				Expect(session.ExitCode()).ToNot(BeZero())
				Expect(session.Err.Contents()).To(ContainSubstring("object doesn't exist"))
			},
			configurations...)
	})
})
