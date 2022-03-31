FROM golang:1.18 AS go

WORKDIR /src
COPY . .

RUN make build

# alpine step
FROM alpine:latest AS alpine

RUN apk --update add ca-certificates

# final step
FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=go /src/bin/app /bin/app

ENTRYPOINT ["/bin/app"]
