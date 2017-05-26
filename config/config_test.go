package config_test

import (
	"bytes"

	. "github.com/cloudfoundry/bosh-gcscli/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreClient configuration", func() {
	Describe("checking that location or storage_class has been set", func() {
		Context("when neither location and storage_class have been set", func() {
			dummyJSONBytes := []byte(`{"bucket_name": "some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			It("defaults to US and MULTI_REGIONAL", func() {
				c, err := NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.Location).To(Equal("us"))
				Expect(c.StorageClass).To(Equal("MULTI_REGIONAL"))
			})
		})

		Context("when storage_class has been set to MULTI_REGIONAL", func() {
			dummyJSONBytes := []byte(`{"storage_class":"MULTI_REGIONAL","bucket_name": "some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			It("defaults to US", func() {
				c, err := NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.Location).To(Equal("us"))
			})
		})

		Context("when storage_class has been set to REGIONAL", func() {
			dummyJSONBytes := []byte(`{"storage_class":"REGIONAL","bucket_name": "some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			It("defaults to us-east1", func() {
				c, err := NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.Location).To(Equal("us-east1"))
			})
		})

		Context("when location has been set to us", func() {
			dummyJSONBytes := []byte(`{"location":"us","bucket_name": "some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			It("defaults to MULTI_REGIONAL", func() {
				c, err := NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.StorageClass).To(Equal("MULTI_REGIONAL"))
			})
		})

		Context("when location has been set to us-west1", func() {
			dummyJSONBytes := []byte(`{"location":"us-west1","bucket_name": "some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			It("defaults to REGIONAL", func() {
				c, err := NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.StorageClass).To(Equal("REGIONAL"))
			})
		})

		DescribeTable("invalid storage_class and location combinations",
			func(dummyJSON string, expected error) {
				dummyJSONBytes := []byte(dummyJSON)
				dummyJSONReader := bytes.NewReader(dummyJSONBytes)

				_, err := NewFromReader(dummyJSONReader)
				Expect(err).To(MatchError(expected))
			},
			Entry("storage_class is MULTI_REGIONAL and location is regional",
				`{"storage_class": "MULTI_REGIONAL", "location":"us-west1","bucket_name": "some-bucket"}`,
				ErrBadLocationStorageClass),
			Entry("storage_class is REGIONAL and location is multi-regional",
				`{"storage_class": "REGIONAL", "location":"us","bucket_name": "some-bucket"}`,
				ErrBadLocationStorageClass),
			Entry("storage_class is unknown",
				`{"storage_class": "asdasdasd","bucket_name": "some-bucket"}`,
				ErrUnknownStorageClass),
			Entry("location is unknown",
				`{"location": "asdasdasd","bucket_name": "some-bucket"}`,
				ErrUnknownLocation))

		Context("when storage_class has been set to MULTI_REGIONAL and location has been set to eu", func() {
			dummyJSONBytes := []byte(`{"storage_class": "MULTI_REGIONAL", "location":"eu","bucket_name": "some-bucket"}`)
			dummyJSONReader := bytes.NewReader(dummyJSONBytes)

			It("uses them", func() {
				c, err := NewFromReader(dummyJSONReader)
				Expect(err).ToNot(HaveOccurred())
				Expect(c.Location).To(Equal("eu"))
				Expect(c.StorageClass).To(Equal("MULTI_REGIONAL"))
			})
		})
	})

	Describe("when bucket is not specified", func() {
		dummyJSONBytes := []byte(`{}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("returns an error", func() {
			_, err := NewFromReader(dummyJSONReader)
			Expect(err).To(MatchError(ErrEmptyBucketName))
		})
	})

	Describe("when bucket is specified", func() {
		dummyJSONBytes := []byte(`{"bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("uses the given bucket", func() {
			c, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.BucketName).To(Equal("some-bucket"))
		})
	})

	Describe("when credentials_source is specified", func() {
		dummyJSONBytes := []byte(`{"credentials_source": "/tmp/foobar.json", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("uses the credentials", func() {
			c, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.CredentialsSource).To(Equal("/tmp/foobar.json"))
		})
	})

	Describe("when credentials_source is not specified", func() {
		dummyJSONBytes := []byte(`{"credentials_source": "", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("uses the Application Default Credentials", func() {
			_, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("when json is invalid", func() {
		dummyJSONBytes := []byte(`{"credentials_source": '`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("returns an error", func() {
			_, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(BeNil())
		})
	})

})
