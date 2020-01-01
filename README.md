# reg-attendee-service

## Overview

A backend service that provides attendee management.

Implemented in go.

Command line arguments
```-config <path-to-config-file> [-migrate-database]```

## Installation

This service uses go modules to provide dependency management, see `go.mod`.

If you place this repository OUTSIDE of your gopath, go build and go test will download
all required dependencies by default. 

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

## for later

- MVP.2
    - metrics for prometheus https://prometheus.io/docs/guides/go-application/
    - support for day guests
    - admin fields handling (subresource w/separate dto only handled by regsys using admin/user auth, invisible fields if user)
    - attendee search by criteria used by regsys
    - optional partner (nick) field for MMC, check for any other missing fields
    - parse Authorization header even when endpoint does not require authorization, so ctx has the user permissions
- later
    - react to context.cancel
    - separate logging target for log output during test runs, so log output can be asserted (and isn't output)
    - security with oauth2 server
        - security using JWT signatures with key in config
        - permissions using JWT
            - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
            - identification as a specific attendee
        - additional acceptance tests for security
