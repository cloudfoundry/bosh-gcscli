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

package config_test

import (
	"bytes"

	. "github.com/cloudfoundry/bosh-gcscli/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BlobstoreClient configuration", func() {
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

	Describe("when credentials_source is 'static' with json_key", func() {
		dummyJSONBytes := []byte(`{"credentials_source": "static", "json_key": "{\"foo\": \"bar\"}", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("uses the credentials", func() {
			c, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.ServiceAccountFile).ToNot(BeEmpty())
		})
	})

	Describe("when credentials_source is 'static' without json_key", func() {
		dummyJSONBytes := []byte(`{"credentials_source": "static", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("returns an error", func() {
			_, err := NewFromReader(dummyJSONReader)
			Expect(err).To(Equal(ErrEmptyServiceAccountFile))
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

	Describe("when encryption_key is specified", func() {
		// encryption_key = []byte{0, 1, 2, ..., 31} as base64
		dummyJSONBytes := []byte(`{"encryption_key": "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("uses the given key", func() {
			c, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(c.EncryptionKey)).To(Equal(32))
			Expect(c.EncryptionKeyEncoded).To(Equal("AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="))
			Expect(c.EncryptionKeySha256).To(Equal("Yw3NKWbEM2aRElRIu7JbT/QSpJxzLbLIq8G4WBvXEN0="))
		})
	})

	Describe("when encryption_key is too long", func() {
		// encryption_key = []byte{0, 1, 2, ..., 31, 32} as base64
		dummyJSONBytes := []byte(`{"encryption_key": "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8g", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("returns an error", func() {
			_, err := NewFromReader(dummyJSONReader)
			Expect(err).To(Equal(ErrWrongLengthEncryptionKey))
		})
	})

	Describe("when encryption_key is malformed", func() {
		// encryption_key is not valid base64
		dummyJSONBytes := []byte(`{"encryption_key": "zzz", "bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("returns an error", func() {
			_, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("when encryption_key is not specified", func() {
		dummyJSONBytes := []byte(`{"bucket_name": "some-bucket"}`)
		dummyJSONReader := bytes.NewReader(dummyJSONBytes)

		It("uses no encryption", func() {
			c, err := NewFromReader(dummyJSONReader)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.EncryptionKey).To(BeNil())
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
