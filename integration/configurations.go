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

package integration

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-gcscli/config"

	. "github.com/onsi/ginkgo/extensions/table"
)

const regionalBucketEnv = "REGIONAL_BUCKET_NAME"
const multiRegionalBucketEnv = "MULTIREGIONAL_BUCKET_NAME"
const publicBucketEnv = "PUBLIC_BUCKET_NAME"

// noBucketMsg is the template used when a BucketEnv's environment variable
// has not been populated.
const noBucketMsg = "environment variable %s expected to contain a valid Google Cloud Storage bucket but was empty"

const getConfigErrMsg = "creating %s configs: %v"

func readBucketEnv(env string) (string, error) {
	bucket := os.Getenv(env)
	if len(bucket) == 0 {
		return "", fmt.Errorf(noBucketMsg, env)
	}
	return bucket, nil
}

func getRegionalConfig() *config.GCSCli {
	var regional string
	var err error

	if regional, err = readBucketEnv(regionalBucketEnv); err != nil {
		panic(fmt.Errorf(getConfigErrMsg, "base", err))
	}

	return &config.GCSCli{BucketName: regional}
}

func getMultiRegionConfig() *config.GCSCli {
	var multiRegional string
	var err error

	if multiRegional, err = readBucketEnv(multiRegionalBucketEnv); err != nil {
		panic(fmt.Errorf(getConfigErrMsg, "base", err))
	}

	return &config.GCSCli{BucketName: multiRegional}
}

func getBaseConfigs() []TableEntry {
	regional := getRegionalConfig()
	multiRegion := getMultiRegionConfig()

	return []TableEntry{
		Entry("Regional bucket, default StorageClass", regional),
		Entry("MultiRegion bucket, default StorageClass", multiRegion),
	}
}

func getPublicConfig() *config.GCSCli {
	public, err := readBucketEnv(publicBucketEnv)
	if err != nil {
		panic(fmt.Errorf(getConfigErrMsg, "public", err))
	}

	return &config.GCSCli{
		BucketName: public,
	}
}

func getInvalidStorageClassConfigs() []TableEntry {
	regional := getRegionalConfig()
	multiRegion := getMultiRegionConfig()

	multiRegion.StorageClass = "REGIONAL"
	regional.StorageClass = "MULTI_REGIONAL"

	return []TableEntry{
		Entry("Multi-Region bucket, regional StorageClass", regional),
		Entry("Regional bucket, Multi-Region StorageClass", multiRegion),
	}
}
