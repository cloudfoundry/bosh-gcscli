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

// GCSBlobstore encapsulates interaction with the GCS blobstore
type GCSBlobstore struct {
	// gcsClient is a pre-configured storage.Client.
	client *storage.Client
	// gcscliConfig is the configuration for interactions with the blobstore
	config *config.GCSCli
}

// getObjectHandle returns a handle to an object at src.
func (client GCSBlobstore) getObjectHandle(src string) *storage.ObjectHandle {
	return client.client.Bucket(client.config.BucketName).Object(src)
}

// New returns a BlobstoreClient configured to operate using the given config
// and client.
//
// The error is returned by s3cli/client convention.
func New(ctx context.Context, gcsClient *storage.Client,
	gcscliConfig *config.GCSCli) (GCSBlobstore, error) {
	if gcsClient == nil {
		return GCSBlobstore{},
			errors.New("nil client causes invalid blobstore")
	}
	if gcscliConfig == nil {
		return GCSBlobstore{},
			errors.New("nil config causes invalid blobstore")
	}
	return GCSBlobstore{gcsClient, gcscliConfig}, nil
}

// Get fetches a blob from the GCS blobstore.
// Destination will be overwritten if it already exists.
func (client GCSBlobstore) Get(src string, dest io.Writer) error {
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
// Put does not retry if upload fails. This is a change from s3cli/client
// which does retry an upload multiple times.
// TODO: implement retry
func (client GCSBlobstore) Put(src io.ReadSeeker, dest string) error {
	remoteWriter := client.getObjectHandle(dest).NewWriter(context.Background())
	if _, err := io.Copy(remoteWriter, src); err != nil {
		log.Println("Upload failed", err.Error())
		return fmt.Errorf("upload failure: %s", err.Error())
	}
	return remoteWriter.Close()
}

// Delete removes a blob from from the GCS blobstore.
//
// If the object does not exist, Delete returns a nil error.
func (client GCSBlobstore) Delete(dest string) error {
	err := client.getObjectHandle(dest).Delete(context.Background())
	if err == storage.ErrObjectNotExist {
		return nil
	}
	return err
}

// Exists checks if a blob exists in the GCS blobstore.
func (client GCSBlobstore) Exists(dest string) (bool, error) {
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
