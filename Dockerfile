FROM golang:1.26 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /src/bin/app /bin/app

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/bin/app"]
