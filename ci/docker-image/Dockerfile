FROM bosh/golang-release

# Packages
RUN DEBIAN_FRONTEND=noninteractive apt-get -y -qq update && apt-get -y -qq install \
  gcc \
  git-core \
  make \
  software-properties-common \
  wget \
  curl

WORKDIR /tmp/docker-build

# Google SDK
ENV GCLOUD_VERSION=257.0.0
ENV GCLOUD_SHA2SUM=2b9eb732206f9c171b6eb9a22083efb98006cc09f1e53f09468626160d3a7cf8

RUN wget https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz \
    -O gcloud_${GCLOUD_VERSION}_linux_amd64.tar.gz && \
    echo "${GCLOUD_SHA2SUM}  gcloud_${GCLOUD_VERSION}_linux_amd64.tar.gz" > gcloud_${GCLOUD_VERSION}_SHA2SUM && \
    shasum -a 256 -cw --status gcloud_${GCLOUD_VERSION}_SHA2SUM && \
    tar xvf gcloud_${GCLOUD_VERSION}_linux_amd64.tar.gz && \
    mv google-cloud-sdk / && cd /google-cloud-sdk  && ./install.sh

ENV PATH=$PATH:/google-cloud-sdk/bin

# Cleanup
RUN rm -rf /tmp/docker-build
