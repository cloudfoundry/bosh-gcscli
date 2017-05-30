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

package config

import (
	"encoding/json"
	"errors"
	"io"
)

// GCSCli represents the configuration for the gcscli
type GCSCli struct {
	// BucketName is the GCS bucket operations will use.
	BucketName string `json:"bucket_name"`
	// CredentialsSource is the location of a Service Account File.
	// If left empty, Application Default Credentials will be used.
	CredentialsSource string `json:"credentials_source"`
	// StorageClass is the type of storage used for objects added to the bucket
	// https://cloud.google.com/storage/docs/storage-classes
	StorageClass string `json:"storage_class"`
}

const (
	defaultRegionalLocation          = "US-EAST1"
	defaultMultiRegionalLocation     = "US"
	defaultRegionalStorageClass      = "REGIONAL"
	defaultMultiRegionalStorageClass = "MULTI_REGIONAL"
)

// ErrEmptyBucketName is returned when a bucket_name in the config is empty
var ErrEmptyBucketName = errors.New("bucket_name must be set")

// getDefaultStorageClass returns the default StorageClass for a given location.
// This takes into account regional/multi-regional incompatibility.
//
// Empty string is returned if the location cannot be matched.
func getDefaultStorageClass(location string) (string, error) {
	if _, ok := GCSMultiRegionalLocations[location]; ok {
		return defaultMultiRegionalStorageClass, nil
	}
	if _, ok := GCSRegionalLocations[location]; ok {
		return defaultRegionalStorageClass, nil
	}
	return "", ErrUnknownLocation
}

// NewFromReader returns the new gcscli configuration struct from the
// contents of the reader.
//
// reader.Read() is expected to return valid JSON.
func NewFromReader(reader io.Reader) (GCSCli, error) {

	dec := json.NewDecoder(reader)
	var c GCSCli
	if err := dec.Decode(&c); err != nil {
		return GCSCli{}, err
	}

	if c.BucketName == "" {
		return GCSCli{}, ErrEmptyBucketName
	}

	return c, nil
}

// FitCompatibleLocation returns whether a provided Location
// can have c.StorageClass objects written to it.
//
// When c.StorageClass is empty, a compatible default is filled in.
//
// nil return value when compatible, otherwise a non-nil explanation.
func (c *GCSCli) FitCompatibleLocation(loc string) error {
	if c.StorageClass == "" {
		var err error
		if c.StorageClass, err = getDefaultStorageClass(loc); err != nil {
			return err
		}
	}

	_, regional := GCSRegionalLocations[loc]
	_, multiRegional := GCSMultiRegionalLocations[loc]
	if !(regional || multiRegional) {
		return ErrUnknownLocation
	}

	if _, ok := GCSStorageClass[c.StorageClass]; !ok {
		return ErrUnknownStorageClass
	}

	return validLocationStorageClass(loc, c.StorageClass)
}
