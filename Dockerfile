FROM golang:1.21 AS go

WORKDIR /src
COPY . .

RUN make build

FROM alpine:latest AS alpine

RUN apk --update add ca-certificates && apk add tzdata

COPY --from=go /src/bin/app /bin/app
ENV TZ=Europe/Rome

ENTRYPOINT ["/bin/app"]
