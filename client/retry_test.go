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

package client

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

type errorTransport struct {
	try int
}

func (e *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	e.try++
	return nil, &net.DNSError{IsTimeout: false, IsTemporary: true}
}

var _ = Describe("RetryTransport", func() {

	Describe("Validate", func() {
		It("It uses a default sleep duration if one isn't provided", func() {
			maxRetries := 1
			et := &errorTransport{}
			client := http.Client{
				Transport: &RetryTransport{
					Base:            et,
					MaxRetries:      maxRetries,
					FirstRetrySleep: 50 * time.Millisecond,
				},
			}
			_, err := client.Get("http://0.0.0.0")
			Expect(et.try).To(Equal(maxRetries + 1))
			Expect(err).To(HaveOccurred())
			Expect(client.Transport.(*RetryTransport).FirstRetrySleep != 0)
		})

		It("It retries the maximum number of times and then fails", func() {
			maxRetries := 3
			et := &errorTransport{}
			client := http.Client{
				Transport: &RetryTransport{
					Base:       et,
					MaxRetries: maxRetries,
				},
			}

			_, err := client.Get("http://0.0.0.0")
			Expect(et.try).To(Equal(maxRetries + 1))
			Expect(err).To(HaveOccurred())
		})

		It("It retries the maximum number of times and then fails", func() {
			maxRetries := 3
			try := 0

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				try++
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer ts.Close()

			client := http.Client{
				Transport: &RetryTransport{
					Base:       http.DefaultTransport,
					MaxRetries: maxRetries,
				},
			}

			res, err := client.Get(ts.URL)
			Expect(try).To(Equal(maxRetries + 1))
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("It retries the maximum number and succeeds on the last try", func() {
			maxRetries := 3
			try := 0

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if try == maxRetries {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
				try++
			}))
			defer ts.Close()

			client := http.Client{
				Transport: &RetryTransport{
					Base:       http.DefaultTransport,
					MaxRetries: maxRetries,
				},
			}

			res, err := client.Get(ts.URL)
			Expect(try).To(Equal(maxRetries + 1))
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})

		It("It retries zero times and succeeds", func() {
			maxRetries := 0
			try := 0

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				try++
			}))
			defer ts.Close()

			client := http.Client{
				Transport: &RetryTransport{
					Base:       http.DefaultTransport,
					MaxRetries: maxRetries,
				},
			}
			res, err := client.Get(ts.URL)
			Expect(try).To(Equal(maxRetries + 1))
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("It retries zero times and succeeds", func() {
			maxRetries := 0
			try := 0

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if try == maxRetries {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
				try++
			}))
			defer ts.Close()

			client := http.Client{
				Transport: &RetryTransport{
					Base:       http.DefaultTransport,
					MaxRetries: maxRetries,
				},
			}
			res, err := client.Get(ts.URL)
			Expect(try).To(Equal(maxRetries + 1))
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})
	})
})
