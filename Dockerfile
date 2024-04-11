FROM golang:1.21.9 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd ./cmd/
RUN CGO_ENABLED=0 go build -o /transmission-exporter ./cmd/transmission-exporter

FROM alpine:3.19.1

RUN apk add --update ca-certificates

COPY --from=build /transmission-exporter /usr/bin/transmission-exporter

RUN chmod 0755 /usr/bin/transmission-exporter

USER nobody

EXPOSE 19091

ENTRYPOINT ["/usr/bin/transmission-exporter"]
