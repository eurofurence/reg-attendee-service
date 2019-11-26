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

## Version History

### v0.1.0

**MVP 1.** Implements attendee resource, userland fields only, with fixed tokens for admin, user, and optionally
staff access. Changes to attendees are historized, except User Comments. All fields are fully validated, permissions 
are checked based on 3 simple groups. This version can be used as a backend of the old regsys and for the new initial 
reg static frontend.

Limitations: 
 - the current fixed-token security model cannot check which user is logged in. This is ok because only the old 
   regsys will know the user / admin tokens. The only token handed out to users must be the staff token.

## TODO

- v0.1.1 (needed for MVP)
    - configurable start time (different for staff and non-staff), refuse with error msg if too early
    - time server endpoint as expected by frontend (different for staff and non-staff, so we can configure it)

## for later

- MVP.2
    - admin fields handling (subresource w/separate dto only handled by regsys using admin/user auth, invisible fields if user)
    - attendee search by criteria used by regsys
    - optional partner (nick) field for MMC, check for any other missing fields
- later
    - react to context.cancel
    - separate logging target for log output during test runs, so log output can be asserted (and isn't output)
    - security with oauth2 server
        - security using JWT signatures with key in config
        - permissions using JWT
            - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
            - identification as a specific attendee
        - additional acceptance tests for security
