FROM golang:1.18 AS go

WORKDIR /src
COPY . .

RUN make build

# alpine step
FROM alpine:latest AS alpine

RUN apk --update add ca-certificates && apk add tzdata

# final step
FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go /src/bin/app /bin/app
ENV TZ=Europe/Rome

ENTRYPOINT ["/bin/app"]
