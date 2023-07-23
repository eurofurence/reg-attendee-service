FROM golang:1 as build

COPY . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" cmd/main.go

RUN chmod 755 main

FROM scratch

COPY --from=build /app/main /main
COPY --from=build /etc/ssl/certs /etc/ssl/certs

# run as an unprivileged unnamed user that has no write permissions on the binary
USER 8877

ENTRYPOINT ["/main"]