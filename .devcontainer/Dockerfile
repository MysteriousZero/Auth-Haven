FROM golang:1.25.1

# ---------------------------------------
# 1. System deps + full locale support
# ---------------------------------------
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    locales locales-all \
    git curl wget unzip bash \
    build-essential bash-completion \
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

# Keep the protobuf compiler explicit and the Go plugins aligned with CI.
RUN PROTOC_VERSION=33.4 && \
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    chmod +x /usr/local/bin/protoc && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Go protobuf + gRPC plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1

# golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Buf CLI
RUN curl -sSL https://github.com/bufbuild/buf/releases/download/v1.26.0/buf-Linux-x86_64 \
    -o /usr/local/bin/buf && chmod +x /usr/local/bin/buf

# -------------------------
# GitHub CLI
# -------------------------
RUN mkdir -p -m 755 /etc/apt/keyrings && \
    wget -qO /etc/apt/keyrings/githubcli-archive-keyring.gpg \
      https://cli.github.com/packages/githubcli-archive-keyring.gpg && \
    chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends gh && \
    rm -rf /var/lib/apt/lists/*

# -------------------------
# 3. Create non-root user
# -------------------------
RUN groupadd --gid 1000 devuser && \
    useradd --uid 1000 --gid 1000 -ms /bin/bash devuser && \
    usermod -aG root devuser && \
    mkdir -p /go/pkg /go/bin /workspace && \
    chown -R devuser:devuser /go /workspace

USER devuser
WORKDIR /workspace

EXPOSE 8080 50051 40000

CMD ["bash"]
