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
	// StorageClass is the style of storage used for the bucket if it needs
	// to be created.
	// https://cloud.google.com/storage/docs/storage-classes
	StorageClass string `json:"storage_class"`
	// Location is the location of the remote bucket if it needs to be
	// created.
	// https://cloud.google.com/storage/docs/bucket-locations
	Location string `json:"location"`
}

const (
	defaultRegionalLocation          = "us-east1"
	defaultMultiRegionalLocation     = "us"
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

// getDefaultStorageClass returns the default Location for a given StorageClass.
// This takes into account regional/multi-regional incompatibility.
//
// Empty string is returned if the location cannot be matched.
func getDefaultLocation(storageClass string) (string, error) {
	if storageClass == regional {
		return defaultRegionalLocation, nil
	} else if _, ok := GCSStorageClass[storageClass]; ok {
		return defaultMultiRegionalLocation, nil
	}
	return "", ErrUnknownStorageClass
}

// NewFromReader returns the new gcscli configuration struct from the
// contents of the reader. Empty fields will be populated with sane defaults.
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

	if c.StorageClass == "" && c.Location == "" {
		c.Location = defaultMultiRegionalLocation
		c.StorageClass = defaultMultiRegionalStorageClass
	}

	var err error
	if c.StorageClass == "" {
		c.StorageClass, err = getDefaultStorageClass(c.Location)
	} else if c.Location == "" {
		c.Location, err = getDefaultLocation(c.StorageClass)
	}
	if err != nil {
		return GCSCli{}, err
	}

	if err := validLocationStorageClass(c.Location, c.StorageClass); err != nil {
		return GCSCli{}, err
	}

	return c, nil
}
