package integration

import (
	"os"

	"github.com/Everlag/gcscli/config"

	"io/ioutil"

	. "github.com/onsi/gomega"
)

// AssertLifecycleWorks tests the main blobstore object lifecycle from
// creation to deletion.
//
// This is using gomega matchers, so it will fail if called outside an
// 'It' test.
func AssertLifecycleWorks(gcsCLIPath string, config *config.GCSCli) {
	expectedString := GenerateRandomString()
	gcsFileName := GenerateRandomString()

	configPath := MakeConfigFile(config)
	defer func() { _ = os.Remove(configPath) }()

	contentFile := MakeContentFile(expectedString)
	defer func() { _ = os.Remove(contentFile) }()

	session, err := RunGCSCLI(gcsCLIPath, configPath,
		"put", contentFile, gcsFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())

	session, err = RunGCSCLI(gcsCLIPath, configPath,
		"exists", gcsFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())
	Expect(session.Err.Contents()).To(MatchRegexp("File '.*' exists in bucket '.*'"))

	tmpLocalFile, err := ioutil.TempFile("", "gcscli-download")
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = os.Remove(tmpLocalFile.Name()) }()
	err = tmpLocalFile.Close()
	Expect(err).ToNot(HaveOccurred())

	session, err = RunGCSCLI(gcsCLIPath, configPath,
		"get", gcsFileName, tmpLocalFile.Name())
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())

	gottenBytes, err := ioutil.ReadFile(tmpLocalFile.Name())
	Expect(err).ToNot(HaveOccurred())
	Expect(string(gottenBytes)).To(Equal(expectedString))

	session, err = RunGCSCLI(gcsCLIPath, configPath,
		"delete", gcsFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(BeZero())

	session, err = RunGCSCLI(gcsCLIPath, configPath,
		"exists", gcsFileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(Equal(3))
	Expect(session.Err.Contents()).To(MatchRegexp("File '.*' does not exist in bucket '.*'"))
}
