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
	Context("with us-east-1, REGIONA (Default Applicaton Credentials) configuration", func() {
		bucketName := os.Getenv(BucketEnv)
		BeforeEach(func() {
			Expect(bucketName).ToNot(BeEmpty(),
				fmt.Sprintf(NoBucketMsg, BucketEnv))
		})

		configurations := []TableEntry{
			Entry("with minimal config", &config.GCSCli{
				BucketName: bucketName,
			}),
		}

		DescribeTable("Blobstore lifecycle works",
			func(config *config.GCSCli) {
				AssertLifecycleWorks(gcsCLIPath, config)
			},
			configurations...)
	})
})
