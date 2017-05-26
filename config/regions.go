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
	"errors"
)

// GCSRegionalLocations are the valid locations for a regional bucket.
var GCSRegionalLocations = map[string]struct{}{
	"us-central1":     struct{}{},
	"us-east1":        struct{}{},
	"us-west1":        struct{}{},
	"us-east4":        struct{}{},
	"europe-west1":    struct{}{},
	"asia-east1":      struct{}{},
	"asia-northeast1": struct{}{},
	"asia-southeast1": struct{}{},
}

// GCSMultiRegionalLocations are the valid locations for
// a multi-regional bucket
var GCSMultiRegionalLocations = map[string]struct{}{
	"asia": struct{}{},
	"eu":   struct{}{},
	"us":   struct{}{},
}

const (
	multiRegional = "MULTI_REGIONAL"
	regional      = "REGIONAL"
	nearline      = "NEARLINE"
	coldline      = "COLDLINE"
)

// GCSStorageClass are the valid storage classes for a bucket.
var GCSStorageClass = map[string]struct{}{
	multiRegional: struct{}{},
	regional:      struct{}{},
	nearline:      struct{}{},
	coldline:      struct{}{},
}

// ErrBadLocationStorageClass is returned when location and storage_class
// cannot be combined
var ErrBadLocationStorageClass = errors.New("incompatible location and storage_class")

// ErrUnknownLocation is returned when a location is chosen that this package
// has no knowledge of.
var ErrUnknownLocation = errors.New("unknown location")

// ErrUnknownStorageClass is returned when a stroage_class is chosen that
// this package has no knowledge of.
var ErrUnknownStorageClass = errors.New("unknown storage_class")

// validDurability returns nil error on valid location-durability combination
// and non-nil explanation on all else.
func validLocationStorageClass(location, storageClass string) error {
	if _, ok := GCSStorageClass[storageClass]; !ok {
		return ErrUnknownStorageClass
	}

	if storageClass == regional {
		if _, ok := GCSMultiRegionalLocations[location]; ok {
			return ErrBadLocationStorageClass
		}
		return nil
	} else if _, ok := GCSStorageClass[storageClass]; ok {
		if _, ok := GCSRegionalLocations[location]; ok {
			return ErrBadLocationStorageClass
		}
		return nil
	} else {
		return ErrUnknownStorageClass
	}
}
