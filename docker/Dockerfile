# syntax=docker/dockerfile:1
#
# Cross-compiled rather than emulated: the build stage always runs on the
# native BUILDPLATFORM and Go targets TARGETARCH itself. CGO is off, so this
# is a plain cross-compile and an arm64 image costs no more than an amd64 one.
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS build

WORKDIR /app

# Dependencies first so a code-only change does not re-download the module
# cache on every build.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/xmpp-bouncer main.go

# distroless:nonroot runs as uid 65532. Nothing here writes to disk: the
# config is env vars and flags, rooms.yaml is read-only, and state lives in
# MariaDB.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/xmpp-bouncer /xmpp-bouncer
# Read at startup by dbWriter.Migrate("migrations"), relative to WORKDIR.
COPY migrations /migrations

WORKDIR /
ENTRYPOINT ["/xmpp-bouncer"]
