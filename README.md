# rexis-go-attendee

## Overview

A backend service that provides attendee management.

Implemented in go.

Command line arguments
```-config <path-to-config-file> [-migrate-database]```

## Installation

This service uses go modules to provide dependency management, see `go.mod`.

If you place this repository OUTSIDE of your gopath, go build and go test will download
all required dependencies by default. 

## TODO:
- functionality MVP.1
    - age check (birthdate validation), incl. tests
    - duplicates check on add attendee, update attendee, incl. tests


```
in attendeectl.validation.go:
// TODO too early or too late birthday - also add to config
```

``` 
in attendeesrv.go:
// TODO duplicate attendee check (this is a business condition) condition is AND of
//     DbQueryHelper.compare("nick", "=", nick)
//     DbQueryHelper.compare("zip", "=", zip)
//     DbQueryHelper.compare("email", "=", email)
//   and for updates
//     DbQueryHelper.compare("id", "<>", id)
```

- later
    - admin fields handling (subresource, but export type&status on get)
    - react to context.cancel
    - separate logging target for log output during test runs, so log output can be asserted (and isn't output)
    - optional partner (nick) field for MMC
    - security with oauth2 server
        - security using JWT signatures with key in config
        - permissions using JWT
            - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
            - identification as a specific attendee
        - additional acceptance tests for security
