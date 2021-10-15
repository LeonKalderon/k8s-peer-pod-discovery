# build stage
FROM golang:1.16-alpine AS builder

# Add the delve
# Compile Delve
RUN go get github.com/go-delve/delve/cmd/dlv

#RUN apk add --no-cache git
RUN mkdir -p /go/src/github.com/LeonKalderon/k8s-peer-pod-discovery
WORKDIR /go/src/github.com/LeonKalderon/k8s-peer-pod-discovery
COPY . .

RUN go get -t ./...
ENV CGO_ENABLED 0
# Build the executable to `/k8s-peer-pod-discovery`. Mark the build as statically linked.
RUN go build -installsuffix 'static' -gcflags "all=-N -l" -o /k8s-peer-pod-discovery

# final stage
FROM alpine:3.7
WORKDIR /app
RUN apk add --no-cache ca-certificates apache2-utils

COPY --from=builder /k8s-peer-pod-discovery /k8s-peer-pod-discovery

EXPOSE 5000

CMD ["/k8s-peer-pod-discovery"]