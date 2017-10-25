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

package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"log"

	"cloud.google.com/go/storage"
	"github.com/cloudfoundry/bosh-gcscli/config"
)

// ErrInvalidROWriteOperation is returned when credentials associated with the
// client disallow an attempted write operation.
var ErrInvalidROWriteOperation = errors.New("the client operates in read only mode. Change 'credentials_source' parameter value ")

// GCSBlobstore encapsulates interaction with the GCS blobstore
type GCSBlobstore struct {
	gcs      *storage.Client
	config   *config.GCSCli
	readOnly bool
}

// validateRemoteConfig determines if the configuration of the client matches
// against the remote configuration and the StorageClass is valid for the location.
//
// If operating in read-only mode, no mutations can be performed
// so the remote bucket location is always compatible.
func (client *GCSBlobstore) validateRemoteConfig() error {
	if client.readOnly {
		return nil
	}

	bucket := client.gcs.Bucket(client.config.BucketName)
	attrs, err := bucket.Attrs(context.Background())
	if err != nil {
		return err
	}
	return client.config.FitCompatibleLocation(attrs.Location)
}

// getObjectHandle returns a handle to an object at src.
func (client GCSBlobstore) getObjectHandle(src string) *storage.ObjectHandle {
	handle := client.gcs.Bucket(client.config.BucketName).Object(src)
	if client.config.EncryptionKey != nil {
		handle = handle.Key(client.config.EncryptionKey)
	}
	return handle
}

// New returns a GCSBlobstore configured to operate using the given config
//
// non-nil error is returned on invalid Client or config. If the configuration
// is incompatible with the GCS bucket, a non-nil error is also returned.
func New(ctx context.Context, cfg *config.GCSCli) (*GCSBlobstore, error) {
	if cfg == nil {
		return nil, errors.New("expected non-nill config object")
	}

	storageClient, readOnly, err := newStorageClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating storage client: %v", err)
	}

	return &GCSBlobstore{gcs: storageClient, config: cfg, readOnly: readOnly}, nil
}

// Get fetches a blob from the GCS blobstore.
// Destination will be overwritten if it already exists.
func (client *GCSBlobstore) Get(src string, dest io.Writer) error {
	remoteReader, err := client.getObjectHandle(src).NewReader(context.Background())
	if err != nil {
		return err
	}
	_, err = io.Copy(dest, remoteReader)
	return err
}

// Put uploads a blob to the GCS blobstore.
// Destination will be overwritten if it already exists.
//
// Put retries retryAttempts times
const retryAttempts = 3

func (client *GCSBlobstore) Put(src io.ReadSeeker, dest string) error {
	if client.readOnly {
		return ErrInvalidROWriteOperation
	}

	if err := client.validateRemoteConfig(); err != nil {
		return err
	}

	var errs []error
	for i := 0; i < retryAttempts; i++ {
		err := client.putOnce(src, dest)
		if err == nil {
			return nil
		}

		errs = append(errs, err)
		log.Printf("upload failed for %s, attempt %d/%d: %v\n", dest, i+1, retryAttempts, err)
	}

	return fmt.Errorf("upload failed for %s after %d attempts: %v", dest, retryAttempts, errs)
}

func (client *GCSBlobstore) putOnce(src io.ReadSeeker, dest string) error {
	remoteWriter := client.getObjectHandle(dest).NewWriter(context.Background())
	remoteWriter.ObjectAttrs.StorageClass = client.config.StorageClass

	if _, err := io.Copy(remoteWriter, src); err != nil {
		remoteWriter.CloseWithError(err)
		return err
	}

	return remoteWriter.Close()
}

// Delete removes a blob from from the GCS blobstore.
//
// If the object does not exist, Delete returns a nil error.
func (client *GCSBlobstore) Delete(dest string) error {
	if client.readOnly {
		return ErrInvalidROWriteOperation
	}

	err := client.getObjectHandle(dest).Delete(context.Background())
	if err == storage.ErrObjectNotExist {
		return nil
	}
	return err
}

// Exists checks if a blob exists in the GCS blobstore.
func (client *GCSBlobstore) Exists(dest string) (bool, error) {
	_, err := client.getObjectHandle(dest).Attrs(context.Background())
	if err == nil {
		log.Printf("File '%s' exists in bucket '%s'\n",
			dest, client.config.BucketName)
		return true, nil
	} else if err == storage.ErrObjectNotExist {
		log.Printf("File '%s' does not exist in bucket '%s'\n",
			dest, client.config.BucketName)
		return false, nil
	}
	return false, err
}
