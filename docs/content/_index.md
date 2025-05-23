---
title: wp-docker-buildx
---

[![Build Status](https://ci.thegeeklab.de/api/badges/thegeeklab/wp-docker-buildx/status.svg)](https://ci.thegeeklab.de/repos/thegeeklab/wp-docker-buildx)
[![Docker Hub](https://img.shields.io/badge/dockerhub-latest-blue.svg?logo=docker&logoColor=white)](https://hub.docker.com/r/thegeeklab/wp-docker-buildx)
[![Quay.io](https://img.shields.io/badge/quay-latest-blue.svg?logo=docker&logoColor=white)](https://quay.io/repository/thegeeklab/wp-docker-buildx)
[![Go Report Card](https://goreportcard.com/badge/github.com/thegeeklab/wp-docker-buildx)](https://goreportcard.com/report/github.com/thegeeklab/wp-docker-buildx)
[![GitHub contributors](https://img.shields.io/github/contributors/thegeeklab/wp-docker-buildx)](https://github.com/thegeeklab/wp-docker-buildx/graphs/contributors)
[![Source: GitHub](https://img.shields.io/badge/source-github-blue.svg?logo=github&logoColor=white)](https://github.com/thegeeklab/wp-docker-buildx)
[![License: Apache-2.0](https://img.shields.io/github/license/thegeeklab/wp-docker-buildx)](https://github.com/thegeeklab/wp-docker-buildx/blob/main/LICENSE)

Woodpecker CI plugin to build multiarch OCI images with buildx.

<!-- prettier-ignore-start -->
<!-- spellchecker-disable -->
{{< toc >}}
<!-- spellchecker-enable -->
<!-- prettier-ignore-end -->

## Usage

{{< hint type=important >}}
Be aware that the this plugin requires [privileged](https://woodpecker-ci.org/docs/usage/workflow-syntax#privileged-mode) capabilities, otherwise the integrated Docker daemon is not able to start.
{{< /hint >}}

```yaml
steps:
  - name: docker
    image: quay.io/thegeeklab/wp-docker-buildx
    privileged: true
    settings:
      username: octocat
      password: random-secret
      repo: octocat/example
      tags: latest
```

### Parameters

<!-- prettier-ignore-start -->
<!-- spellchecker-disable -->
{{< propertylist name=wp-docker-buildx.data sort=name >}}
<!-- spellchecker-enable -->
<!-- prettier-ignore-end -->

### Examples

#### Push to other registries than DockerHub

If the created image is to be pushed to registries other than the default DockerHub, it is necessary to set `registry` and `repo` as fully-qualified name.

**GHCR:**

```yaml
steps:
  - name: docker
    image: quay.io/thegeeklab/wp-docker-buildx
    privileged: true
    settings:
      registry: ghcr.io
      username: octocat
      password: secret-access-token
      repo: ghcr.io/octocat/example
      tags: latest
```

**AWS ECR:**

```yaml
steps:
  - name: docker
    image: quay.io/thegeeklab/wp-docker-buildx
    privileged: true
    settings:
      environment:
        AWS_ACCESS_KEY_ID:
          from_secret: aws_access_key_id
        AWS_SECRET_ACCESS_KEY:
          from_secret: aws_secret_access_key
      registry: <account_id>.dkr.ecr.<region>.amazonaws.com
      repo: <account_id>.dkr.ecr.<region>.amazonaws.com/octocat/example
      tags: latest
```

## Build

Build the binary with the following command:

```shell
make build
```

Build the container image with the following command:

```shell
docker build --file Containerfile.multiarch --tag thegeeklab/wp-docker-buildx .
```

## Test

```shell
docker run --rm \
  -e PLUGIN_TAG=latest \
  -e PLUGIN_REPO=octocat/hello-world \
  -e CI_COMMIT_SHA=00000000 \
  -v $(pwd)/testdata:/build:z \
  -w /build \
  --privileged \
  thegeeklab/wp-docker-buildx --dry-run
```
