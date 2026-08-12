# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.1
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

ARG TARGETARCH

COPY . .
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./cmd/server && \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/migrate ./cmd/migrate

FROM alpine:3.22 AS final

RUN apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser
WORKDIR /app

COPY --from=build /bin/server /bin/migrate /bin/
COPY migrations /app/migrations

EXPOSE 8080 50051

ENTRYPOINT [ "/bin/server" ]
