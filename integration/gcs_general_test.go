package integration

import (
	"fmt"
	"os"

	"github.com/Everlag/gcscli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const BucketEnv = "BUCKET_NAME"

// NoBucketMsg is the template used when BucketEnv's environment variable
// has not been populated.
const NoBucketMsg = "environment variable %s expected to contain a valid Google Cloud Storage bucket but was empty"

var _ = Describe("Integration", func() {
	Context("general (Default Applicaton Credentials) configuration", func() {
		bucketName := os.Getenv(BucketEnv)

		var ctx AssertContext
		BeforeEach(func() {
			Expect(bucketName).ToNot(BeEmpty(),
				fmt.Sprintf(NoBucketMsg, BucketEnv))

			ctx = NewAssertContext()
		})
		AfterEach(func() {
			ctx.Cleanup()
		})

		configurations := []TableEntry{
			Entry("with minimal config", &config.GCSCli{
				BucketName: bucketName,
			}),
		}

		DescribeTable("Blobstore lifecycle works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertLifecycleWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Invalid Delete works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertDeleteNonexistentWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Multipart Put works",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertMultipartPutWorks(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Invalid Put should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertOnPutFails(gcsCLIPath, ctx)
			},
			configurations...)

		DescribeTable("Invalid Get should fail",
			func(config *config.GCSCli) {
				ctx.AddConfig(config)
				AssertGetNonexistentFails(gcsCLIPath, ctx)
			},
			configurations...)
	})
})
