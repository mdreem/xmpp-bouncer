FROM golang:1.19-buster as build

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o bin/xmpp-bouncer main.go

FROM gcr.io/distroless/static-debian11@sha256:5bda5a482de9b13160c061a3fd8665eecf8e35f5ed78e324dafdcdb11cf6d5da
COPY --from=build /app/bin/xmpp-bouncer /xmpp-bouncer
ENTRYPOINT ["/xmpp-bouncer"]
