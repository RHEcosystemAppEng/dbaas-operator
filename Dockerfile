# Build the manager binary
FROM registry.access.redhat.com/ubi8:8.5 AS builder

# Set go version
ARG RUNTIME_VERSION=1.16.15

RUN curl -fsSLo /tmp/go.tgz https://golang.org/dl/go${RUNTIME_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tgz && \
    ln -s ../go/bin/go /usr/local/bin/go && \
    rm /tmp/go.tgz && \
    go version

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Build the operator image
FROM registry.access.redhat.com/ubi8-minimal:8.5

COPY LICENSE /licenses/LICENSE
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
