FROM bosh/golang-release

ENV DEBIAN_FRONTEND "noninteractive"

RUN apt-get update -y \
    && apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        git \
        wget

RUN apt-get update -y \
    && apt-get -y install --no-install-recommends \
      apt-transport-https \
      ca-certificates \
      gnupg \
    && export GCLOUD_KEYFILE="/usr/share/keyrings/cloud.google.gpg" \
    && echo "deb [signed-by=${GCLOUD_KEYFILE}] https://packages.cloud.google.com/apt cloud-sdk main" \
    | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list \
    && curl -fSL https://packages.cloud.google.com/apt/doc/apt-key.gpg \
    | apt-key --keyring "${GCLOUD_KEYFILE}" add - \
    && apt-get update -y \
    && apt-get install -y --no-install-recommends \
      google-cloud-cli
