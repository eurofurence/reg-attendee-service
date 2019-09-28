# rexis-go-attendee

## Overview

A backend service that provides attendee management.

Implemented in go.

Command line arguments
```-config <path-to-config-file>```

## Installation

This service uses go modules to provide dependency management, see `go.mod`.

If you place this repository OUTSIDE of your gopath, go build and go test will download
all required dependencies by default. 

## TODO:
- functionality MVP.1
    - remaining field validations (email pattern, flag/options/pkg logic, tshirt sizes conf and validation)
- security with oauth2 server MVP.2
    - security using JWT signatures with key in config
    - permissions using JWT
        - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
        - identification as a specific attendee
    - acceptance tests for security
- later
    - admin fields handling (subresource, but export type&status on get)
    - react to context.cancel
    - separate logging target for log output during test runs, so log output can be asserted (and isn't output)
    - optional partner (nick) field for MMC
