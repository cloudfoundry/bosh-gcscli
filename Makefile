# Copyright 2017 Google Inc.
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#    http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
SHELL = bash
default: test-int

# build the binary
build:
	go install

# Fetch base dependencies as well as testing packages
get-deps:

# Cleans up directory and source code with gofmt
clean:
	go clean ./...

# Run gofmt on all code
fmt:
	gofmt -l -w $$(ls -d */ | grep -v vendor)

# Run linter with non-strict checking
lint:
	@if ! command -v golangci-lint &> /dev/null; then \
	  go_bin="$(go env GOPATH)/bin"; \
	  export PATH=${go_bin}:${PATH}; \
	  go install -v github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi; \
	golangci-lint run ./...

# Generate a $StorageClass.lock which contains our bucket name
# used for testing. Buckets must be unique among all in GCS,
# we cannot simply hardcode a bucket.
.PHONY: FORCE
regional.lock:
	@test -s "regional.lock" || \
	{ echo -n bosh-gcs; \
	cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 40 | head -n 1 ;} > regional.lock

# Create a bucket using the name located in $StorageClass.lock with
# a sane location.
regional-bucket: regional.lock
	@gsutil ls | grep -q "$$(cat regional.lock)"; if [ $$? -ne 0 ]; then \
		gsutil mb -c REGIONAL -l us-east1 "gs://$$(cat regional.lock)"; \
	fi

.PHONY: FORCE
multiregional.lock:
	@test -s "multiregional.lock" || \
	{ echo -n bosh-gcs; \
	cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 40 | head -n 1 ;} > multiregional.lock

multiregional-bucket: multiregional.lock
	@gsutil ls | grep -q "$$(cat multiregional.lock)"; if [ $$? -ne 0 ]; then \
		gsutil mb -c MULTI_REGIONAL -l us "gs://$$(cat multiregional.lock)"; \
	fi

.PHONY: FORCE
public.lock:
	@test -s "public.lock" || \
	{ echo -n bosh-gcs; \
	cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 40 | head -n 1 ;} > public.lock


public-bucket: public.lock
	@gsutil ls | grep -q "$$(cat public.lock)"; if [ $$? -ne 0 ]; then \
		gsutil mb -c MULTI_REGIONAL -l us "gs://$$(cat public.lock)" && \
		gsutil iam ch allUsers:legacyObjectReader "gs://$$(cat public.lock)" && \
		gsutil iam ch allUsers:legacyBucketReader "gs://$$(cat public.lock)" && \
		echo "waiting for IAM to propagate" && \
		until curl -s \
			"https://storage.googleapis.com/$$(cat public.lock)/non-existent" \
			| grep -q "NoSuchKey"; do sleep 1; done; \
	fi

# Create all buckets necessary for the test.
prep-gcs: regional-bucket multiregional-bucket public-bucket

# Remove all buckets listed in $StorageClass.lock files.
clean-gcs:
	@test -s "multiregional.lock" && test -s "regional.lock" && test -s "public.lock"
	@gsutil rm "gs://$$(cat regional.lock)/*" || true
	@gsutil rb "gs://$$(cat regional.lock)"
	@rm regional.lock
	@gsutil rm "gs://$$(cat multiregional.lock)/*" || true
	@gsutil rb "gs://$$(cat multiregional.lock)"
	@rm multiregional.lock
	@gsutil rm "gs://$$(cat public.lock)/*" || true
	@gsutil rb "gs://$$(cat public.lock)"
	@rm public.lock

# Perform only unit tests
test-unit: get-deps clean fmt lint build
	go run github.com/onsi/ginkgo/v2/ginkgo run -r --skip-package integration

.PHONY:
check-int-env:
ifndef GOOGLE_SERVICE_ACCOUNT
	$(error environment variable GOOGLE_SERVICE_ACCOUNT is undefined)
endif

# Perform all tests, including integration tests.
test-int: get-deps clean fmt lint build prep-gcs check-int-env
	 export MULTIREGIONAL_BUCKET_NAME="$$(cat multiregional.lock)" && \
	 export REGIONAL_BUCKET_NAME="$$(cat regional.lock)" && \
	 export PUBLIC_BUCKET_NAME="$$(cat public.lock)" && \
	 go run github.com/onsi/ginkgo/v2/ginkgo run -r

# Perform all non-long tests, including integration tests.
test-fast-int: get-deps clean fmt lint build prep-gcs check-int-env
	 export MULTIREGIONAL_BUCKET_NAME="$$(cat multiregional.lock)" && \
	 export REGIONAL_BUCKET_NAME="$$(cat regional.lock)" && \
	 export PUBLIC_BUCKET_NAME="$$(cat public.lock)" && \
	 export SKIP_LONG_TESTS="yes" && \
	 go run github.com/onsi/ginkgo/v2/ginkgo run -r

help:
	 @echo "common developer commands:"
	 @echo "  get-deps: fetch developer dependencies"
	 @echo "  fmt: run gofmt on the codebase"
	 @echo "  clean: run go clean on the codebase"
	 @echo "  lint: run go lint on the codebase"
	 @echo ""
	 @echo "common testing commands:"
	 @echo "  prep-gcs: create external GCS buckets needed for integration testing"
	 @echo "  clean-gcs: remove external GCS buckets"
	 @echo "  test-fast-int: run an reduced integration test suite (presubmit)"
	 @echo "  test-int: run the full integration test (CI only)"
	 @echo ""
	 @echo "expected environment variables:"
	 @echo "  GOOGLE_SERVICE_ACCOUNT=contents of a JSON service account key"
