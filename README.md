# rexis-go-attendee

## Overview

A backend service that provides attendee management.

Implemented in go.

Command line arguments
```-config <path-to-config-file>```

## Installation

```
go get gopkg.in/yaml.v2
go get github.com/gorilla/mux
go get github.com/go-http-utils/headers
go get github.com/jinzhu/gorm
go get github.com/go-sql-driver/mysql
go get github.com/stretchr/testify
```

TODO:
- request level acceptance tests
- integration tests 
- contract tests
- unit tests (mapping logic, ...)
- introduce context.Context everywhere (OMG more clutter)
- request logging
- separate logging target for log output during test runs, so log output can be asserted (and isn't output)
- assign request id and return in case of errors
- log request id everywhere, log severity everywhere, log format -> wrap logging
- security using JWT signatures with key in config
- permissions using JWT
    - viewAttendees, changeAttendees, viewAttendeeAdmininfo, changeAttendeeAdmininfo rights
    - identification as a specific attendee
- acceptance tests for security
- maintain change history in DB
- admin fields handling (subresource, but export type&status on get)
- remaining field validations (email pattern, flag/options/pkg logic)
