FROM golang:1.26 AS go

WORKDIR /src

COPY . .

RUN go mod download

RUN make build

FROM alpine:latest AS alpine

RUN apk --no-cache --update add ca-certificates tzdata

COPY --from=go /src/bin/app /bin/app
ENV TZ=Europe/Rome

ENTRYPOINT ["/bin/app"]
