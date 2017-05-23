package client

import (
	"context"

	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
	"github.com/Everlag/gcscli/config"
)

const uaString = "gcscli"

// NewSDK returns context and client necessary to instantiate a client
// based off of the provided configuration.
func NewSDK(c config.GCSCli) (context.Context, *storage.Client, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error
	ua := option.WithUserAgent(uaString)
	if c.CredentialsSource == "" {
		client, err = storage.NewClient(ctx, ua)
	} else {
		client, err = storage.NewClient(ctx, ua,
			option.WithServiceAccountFile(c.CredentialsSource))
	}
	return ctx, client, err
}
