# BUILD STAGE
FROM golang:1.14 AS builder

WORKDIR /adoption-metrics

# copy go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# copy source files
COPY main.go main.go

RUN echo "+ Generating metrics binary"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o "metrics" ./main.go

FROM alpine:3.9

COPY --from=builder /adoption-metrics/metrics /usr/bin/
COPY config.yml /config.yml

CMD ["/usr/bin/metrics"]

EXPOSE 9104
      