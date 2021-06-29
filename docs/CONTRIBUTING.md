# Contributing


## Prerequisites

- [Go](https://golang.org/), tested on go1.16 darwin/amd64 and linux/arm
- [Git](https://git-scm.com/)
- GNU Make
- [goreleaser](https://github.com/goreleaser/goreleaser) for local release test before tagging release
- [grafana/json-schema-docs](https://github.com/grafana/json-schema-docs) to generate schema docs

## Build

```console
$ git clone https://github.com/soracom/soratun
$ cd soratun
$ make
```

If you update configuration file (`arc.json`) format, please update relevant [JSON
schema](https://json-schema.org/) ([English](schema/soratun-config.en.schema.json) / [Japanese](schema/soratun-config.ja.schema.json)) and generate new
documents for users.

```console
$ make json-schema-docs
```

## Test

### WireGuard Integration Test

#### Build a container for testing

```
$ make integration-test-container
```

#### How to run it

```
$ make run-integration-test-container
$ make integration-test
```

#### Push a testing container to GitHub Registry for CI

```
$ make integration-test-container
$ make test-docker-container-push DOCKER_USER=${your_github_user_name} DOCKER_PSWD_FILE=/path/to/your/github/token/file
```

See also: https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token

This personal access token must be granted the appropriate permission to push the container.

## Release

Tagging a commit and push it to the repo will create a new release with configured [GitHub action](https://github.com/soracom/soratun/actions), under the [Releases](https://github.com/soracom/soratun/releases/) section.

```console
$ git tag v0.16.0
$ git push --tags
```

Before pushing your new tag, please test it locally with:

```console
$ make snapshot
```

If no error, the action should work.

## Debugging Tips

- Set `SORACOM_VERBOSE=1` environment variable to see API requests and responses.
