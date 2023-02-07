# reg-attendee-service

<img src="https://github.com/eurofurence/reg-attendee-service/actions/workflows/go.yml/badge.svg" alt="test status"/>

## Overview

A backend service that provides attendee management.

Implemented in go.

Command line arguments
```
-config <path-to-config-file> [-migrate-database] [-ecs-json-logging]
```

## Installation

This service uses go modules to provide dependency management, see `go.mod`.

If you place this repository outside your GOPATH, build and test runs will download all required 
dependencies by default. 

## Running on localhost

Copy the configuration template from `docs/config-template.yaml` to `./config.yaml`. This will set you up
for operation with an in-memory database and sensible defaults.

Build using `go build cmd/main.go`.

Then run `./main -config config.yaml -migrate-database`.

## Installation on the server

See `install.sh`. This assumes a current build, and a valid configuration template in specific filenames.

## Test Coverage

In order to collect full test coverage, set go tool arguments to `-covermode=atomic -coverpkg=./internal/...`,
or manually run
```
go test -covermode=atomic -coverpkg=./internal/... ./...
```

## Contract Testing

This microservice uses [pact-go](https://github.com/pact-foundation/pact-go#installation) for contract tests.

Before you can run the contract tests in this repository, you need to run the client side contract tests
in the [reg-attendee-transferclient](https://github.com/eurofurence/reg-attendee-transferclient) to generate
the contract specification. 

You are expected to clone that repository into a directory called `reg-attendee-transferclient`
right next to this repository. If you wish to place your contract specs somewhere else, simply change the
path or URL in `test/contract/producer/setup_ctr_test.go`.

## Open Issues and Ideas

We track open issues as GitHub issues on this repository once it becomes clear what exactly needs to be done.

### plans for later

- self cancellation if no payments made and before a grace period
- more fine-grained permissions using JWT
  - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
- container build and associated changes
