# reg-attendee-service

<img src="https://github.com/eurofurence/reg-attendee-service/actions/workflows/go.yml/badge.svg" alt="test status"/>
<img src="https://github.com/eurofurence/reg-attendee-service/actions/workflows/codeql-analysis.yml/badge.svg" alt="code quality status"/>

## Overview

A backend service that provides attendee management.

Implemented in go.

Command line arguments
```
-config <path-to-config-file> [-migrate-database] [-ecs-json-logging]
```

## Installation

This service uses go modules to provide dependency management, see `go.mod`.

If you place this repository outside of your GOPATH, build and test runs will download all required 
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

## Version History

### v0.1.0

**MVP 1.** Implements attendee resource, userland fields only, with fixed tokens for admin, user, and optionally
staff access. Changes to attendees are historized, except User Comments. All fields are fully validated, permissions 
are checked based on 3 simple groups. This version can be used as a backend of the old regsys and for the new initial 
reg static frontend.

Limitations: 
 - the current fixed-token security model cannot check which user is logged in. This is ok because only the old 
   regsys will know the user / admin tokens. The only token handed out to users must be the staff token.

### v0.1.1

**MVP 1.1** Implements a countdown resource with configurable start time for public registration. Before that time,
no registrations are accepted. The countdown resource response is formatted as expected by the frontend.

Limitations: 
 - the current fixed-token security model cannot check which user is logged in. This is ok because only the old 
   regsys will know the user / admin tokens. The only token handed out to users must be the staff token.
 - before the configured registration start time, even admin or staff authenticated users will not be able to
   register because the endpoint does not honor a supplied Authorization header at all. This is ok because
   currently we use a separate installation for staff reg with a secret link.

### v0.2.0

**MVP 2.** The absolute minimum needed for EF and MMC reg to work.

 - ✅ finalized open api v3 spec
 - ✅ implements admin fields handling
 - ✅ implements status transitions
 - ✅ includes an openapi spec
 - 🚧 talks to payment service as appropriate (with contract tests)
 - 🚧 talks to mail service as appropriate (with contract tests)
 - ✅ obtains IDP tokens from the cookies set by the auth service, as well as fixed token security for backend requests
 - ✅ auth header and tokens are honored for all requests, even the ones that do not require authorization
 - 🚧 fields for MMC have been added as well (partner, ...) 
 - ✅ day guests are supported simply via the package subsystem 
 - 🚧 guests are supported as an admin only flag which will cause the system to assign 0 dues
 - ✅ implements a general request timeout and panic handling
 - ❌ no search functionality implemented yet
 - ❌ no bans support implemented at this point
 - ❌ no manual dues support implemented yet
- 🚧 key_deposit/key_received/sponsor_items flag are supported as additional-info not implement yet
- 🚧 track who (subject, if set) performed a status change

### for later

- ❌ more fine-grained permissions using JWT
  - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
- ❌ container build and associated changes

