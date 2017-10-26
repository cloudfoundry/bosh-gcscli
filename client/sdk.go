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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/option"

	"net/http"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/storage"
	"github.com/cloudfoundry/bosh-gcscli/config"
)

const uaString = "bosh-gcscli"

func newStorageClient(ctx context.Context, cfg *config.GCSCli) (*storage.Client, bool, error) {
	// default to a read-only client
	readOnly := true
	opt := option.WithHTTPClient(http.DefaultClient)

	switch cfg.CredentialsSource {
	case config.NoneCredentialsSource:
		// no-op
	case config.DefaultCredentialsSource:
		// attempt to load the application default credentials
		if tokenSource, err := getDefaultTokenSource(ctx); err == nil {
			opt = option.WithTokenSource(tokenSource)
			readOnly = false
		}
	case config.ServiceAccountFileCredentialsSource:
		if token, err := google.JWTConfigFromJSON([]byte(cfg.ServiceAccountFile), storage.ScopeFullControl); err == nil {
			opt = option.WithTokenSource(token.TokenSource(ctx))
			readOnly = false
		}
	default:
		return nil, false, errors.New("unknown credentials_source in configuration")
	}

	gcs, err := storage.NewClient(ctx, option.WithUserAgent(uaString), opt)

	return gcs, readOnly, err
}

// If we're on GCE then google.DefaultTokenSource may return the default service account.
// If that account doesn't have storage.ScopeFullControl then we do not want it.
func useDefaultTokenSource() bool {
	if !metadata.OnGCE() {
		return true
	}

	scopes, err := metadata.Scopes("")
	if err != nil {
		// no default service account to use
		return false
	}

	for _, scope := range scopes {
		if scope == storage.ScopeFullControl {
			// default service account has proper scope, it's fine to use it
			return true
		}
	}

	return false
}

func getDefaultTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	if !useDefaultTokenSource() {
		return nil, errors.New("refusing to use default token source")
	}

	return google.DefaultTokenSource(ctx, storage.ScopeFullControl)
}
