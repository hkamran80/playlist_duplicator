# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -o /playlist-duplicator

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /playlist-duplicator /playlist-duplicator

USER nonroot:nonroot

ENTRYPOINT ["/playlist-duplicator"]