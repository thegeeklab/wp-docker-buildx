---
title: wp-docker-buildx
---

[![Build Status](https://img.shields.io/wp/build/thegeeklab/wp-docker-buildx?logo=wp&server=https%3A%2F%2Fwp.thegeeklab.de)](https://wp.thegeeklab.de/thegeeklab/wp-docker-buildx)
[![Docker Hub](https://img.shields.io/badge/dockerhub-latest-blue.svg?logo=docker&logoColor=white)](https://hub.docker.com/r/thegeeklab/wp-docker-buildx)
[![Quay.io](https://img.shields.io/badge/quay-latest-blue.svg?logo=docker&logoColor=white)](https://quay.io/repository/thegeeklab/wp-docker-buildx)
[![GitHub contributors](https://img.shields.io/github/contributors/thegeeklab/wp-docker-buildx)](https://github.com/thegeeklab/wp-docker-buildx/graphs/contributors)
[![Source: GitHub](https://img.shields.io/badge/source-github-blue.svg?logo=github&logoColor=white)](https://github.com/thegeeklab/wp-docker-buildx)
[![License: MIT](https://img.shields.io/github/license/thegeeklab/wp-docker-buildx)](https://github.com/thegeeklab/wp-docker-buildx/blob/main/LICENSE)

Woodpecker CI plugin to build multiarch OCI images with buildx.

<!-- prettier-ignore-start -->
<!-- spellchecker-disable -->
{{< toc >}}
<!-- spellchecker-enable -->
<!-- prettier-ignore-end -->

## Usage

{{< hint type=important >}}
Be aware that the this plugin requires [privileged](https://docs.wp.io/pipeline/docker/syntax/steps/#privileged-mode) capabilities, otherwise the integrated Docker daemon is not able to start.
{{< /hint >}}

```yaml
kind: pipeline
name: default

steps:
  - name: docker
    image: thegeeklab/wp-docker-buildx:23
    privileged: true
    settings:
      username: octocat
      password: secure
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
kind: pipeline
name: default

steps:
  - name: docker
    image: thegeeklab/wp-docker-buildx:23
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
kind: pipeline
name: default

steps:
  - name: docker
    image: thegeeklab/wp-docker-buildx:23
    privileged: true
    environment:
      AWS_ACCESS_KEY_ID:
        from_secret: aws_access_key_id
      AWS_SECRET_ACCESS_KEY:
        from_secret: aws_secret_access_key
    settings:
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
  -v $(pwd):/build:z \
  -w /build \
  --privileged \
  thegeeklab/wp-docker-buildx --dry-run
```
