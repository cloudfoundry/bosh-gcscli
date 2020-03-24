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
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	defaultFirstRetrySleep = 50 * time.Millisecond
)

// RequestModifier is a function that will modify the request before it is made
type RequestModifier func(req *http.Request)

// RetryTransport is a function that will retry failed HTTP connections up to
// a maximum number of times.
type RetryTransport struct {
	MaxRetries      int
	FirstRetrySleep time.Duration
	Base            http.RoundTripper
	RequestModifier RequestModifier
}

func (rt *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.try(req)
}

func (rt *RetryTransport) try(req *http.Request) (resp *http.Response, err error) {
	if rt.FirstRetrySleep == 0 {
		rt.FirstRetrySleep = defaultFirstRetrySleep
	}

	var body []byte

	if rt.RequestModifier != nil {
		rt.RequestModifier(req)
	}

	// Save the req body for future retries as it will be read and closed
	// by Base.RoundTrip.
	if req.Body != nil {
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return
		}
	}

	for try := 0; try <= rt.MaxRetries; try++ {
		r := bytes.NewReader(body)
		req.Body = ioutil.NopCloser(r)
		resp, err = rt.Base.RoundTrip(req)

		sleep := func() {
			d := rt.FirstRetrySleep << uint64(try)
			log.Printf("RetryTransport: Retrying request (%d/%d) after %s", try, rt.MaxRetries, d)
			time.Sleep(d)
		}

		// Retry on net.Error
		switch err.(type) {
		case net.Error:
			if !err.(net.Error).Temporary() {
				return
			}
			sleep()
			continue
		case error:
			return
		}

		// Retry on status code >= 500
		if resp.StatusCode >= 500 {
			sleep()
			continue
		}
		return
	}
	return
}
