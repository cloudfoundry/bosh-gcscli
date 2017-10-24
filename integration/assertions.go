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
	"io"
	"os"

	"github.com/cloudfoundry/bosh-gcscli/client"

	"crypto/rand"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// NoLongEnv must be set in the environment
// to enable skipping long running tests
const NoLongEnv = "SKIP_LONG_TESTS"

// NoLongMsg is the template used when BucketNoLongEnv's environment variable
// has not been populated.
const NoLongMsg = "environment variable %s filled, skipping long test"

// AssertLifecycleWorks tests the main blobstore object lifecycle from
// creation to deletion.
//
// This is using gomega matchers, so it will fail if called outside an
// 'It' test.
func AssertLifecycleWorks(gcsCLIPath string, ctx AssertContext) {
	session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"put", ctx.ContentFile, ctx.GCSFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())

	session, err = RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"exists", ctx.GCSFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())
	Expect(session.Err.Contents()).To(MatchRegexp("File '.*' exists in bucket '.*'"))

	tmpLocalFile, err := ioutil.TempFile("", "gcscli-download")
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = os.Remove(tmpLocalFile.Name()) }()
	err = tmpLocalFile.Close()
	Expect(err).ToNot(HaveOccurred())

	session, err = RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"get", ctx.GCSFileName, tmpLocalFile.Name())
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())

	gottenBytes, err := ioutil.ReadFile(tmpLocalFile.Name())
	Expect(err).ToNot(HaveOccurred())
	Expect(string(gottenBytes)).To(Equal(ctx.ExpectedString))

	session, err = RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"delete", ctx.GCSFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())

	session, err = RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"exists", ctx.GCSFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(Equal(3))
	Expect(session.Err.Contents()).To(MatchRegexp("File '.*' does not exist in bucket '.*'"))
}

// AssertDeleteNonexistentWorks tests that attempting to delete a non-existent
// file will be silently ignored.
func AssertDeleteNonexistentWorks(gcsCLIPath string, ctx AssertContext) {
	session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"delete", ctx.GCSFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())
}

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

const twoGB = 1024 * 1024 * 1024 * 2

// AssertMultipartPutWorks tests that attempting to upload a large,
// multipart blob succeeds.
func AssertMultipartPutWorks(gcsCLIPath string, ctx AssertContext) {
	if os.Getenv(NoLongEnv) != "" {
		Skip(fmt.Sprintf(NoLongMsg, NoLongEnv))
	}

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

// AssertBrokenSourcePutFails tests that a broken upload will cause a failure
func AssertBrokenSourcePutFails(gcsCLIPath string, ctx AssertContext) {
	_, gcsClient, err := client.NewSDK(*ctx.Config)
	Expect(err).ToNot(HaveOccurred())
	blobstoreClient, err := client.New(context.Background(),
		gcsClient, ctx.Config)
	Expect(err).ToNot(HaveOccurred())

	err = blobstoreClient.Put(&badReadSeeker{}, ctx.GCSFileName)
	Expect(err).To(HaveOccurred())
}

// AssertGetNonexistentFails tests that attempting to get a non-existent
// object will fail.
func AssertGetNonexistentFails(gcsCLIPath string, ctx AssertContext) {
	session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"get", ctx.GCSFileName, "/dev/null")
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).ToNot(BeZero())
	Expect(session.Err.Contents()).To(ContainSubstring("object doesn't exist"))
}

// AssertPutFails tests that whatever context is passed will cause a put
// operation to fail.
func AssertPutFails(gcsCLIPath string, ctx AssertContext) {
	session, err := RunGCSCLI(gcsCLIPath, ctx.ConfigPath,
		"put", ctx.ContentFile, ctx.GCSFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).ToNot(BeZero())
}
