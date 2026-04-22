FROM golang:1.26.2

# ---------------------------------------
# 1. System deps + full locale support
# ---------------------------------------
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    locales locales-all \
    git curl wget unzip bash \
    build-essential bash-completion \
    protobuf-compiler \
    && sed -i 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen \
    && locale-gen en_US.UTF-8 \
    && update-locale LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8 \
    && rm -rf /var/lib/apt/lists/*

ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    GO111MODULE=on \
    CGO_ENABLED=1 \
    GOPATH=/go \
    PATH=/go/bin:/usr/local/go/bin:$PATH

# -------------------------
# 2. Dev tooling
# -------------------------

# Delve debugger
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Protoc latest
RUN PROTOC_VERSION=25.1 && \
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    chmod +x /usr/local/bin/protoc && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Go protobuf + gRPC plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b /go/bin

# Buf CLI
RUN curl -sSL https://github.com/bufbuild/buf/releases/download/v1.26.0/buf-Linux-x86_64 \
    -o /usr/local/bin/buf && chmod +x /usr/local/bin/buf

# -------------------------
# 3. Create non-root user
# -------------------------
# Pin docker group to GID 0 so it matches the host docker socket GID.
# Since the socket is owned by root:root on the host, we add devuser to the root group.
RUN useradd -ms /bin/bash devuser && \
    usermod -aG root devuser && \
    mkdir -p /go/pkg /go/bin /workspace && \
    chown -R devuser:devuser /go /workspace

USER devuser
WORKDIR /workspace

EXPOSE 8080 40000

CMD ["bash"]