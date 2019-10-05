# build app
FROM golang:alpine as build-env

RUN apk --no-cache add build-base git gcc ca-certificates
COPY go.mod main.go pagetemplate.go /src/
RUN cd /src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main

# create final image
FROM scratch

COPY --from=build-env /src/main /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/main"]