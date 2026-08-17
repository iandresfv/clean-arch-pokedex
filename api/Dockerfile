# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Build stage
#
# The Go toolchain and the source live only here. The runtime stage copies the
# compiled binary and nothing else, so neither ships to production.
# ---------------------------------------------------------------------------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Dependencies are copied and downloaded before the source so that this layer is
# cached and only re-runs when go.mod or go.sum changes.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH

# CGO_ENABLED=0 produces a statically linked binary with no dynamic
# dependencies. Without it the binary links the host's libc and will not run on
# a distroless or scratch image.
#
# -s -w strip the symbol table and DWARF debug info, roughly halving the binary.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/server ./cmd/server/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/seeder ./cmd/seeder/

# ---------------------------------------------------------------------------
# Runtime stage
#
# distroless/static rather than alpine: with a static binary there is nothing to
# install, no package manager to patch, and no shell for an attacker to reach.
# Alpine's value is having a shell, which is exactly what is not wanted here.
# ---------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /out/seeder /app/seeder
COPY --from=build /src/migrations /app/migrations

# The nonroot tag runs as uid 65532. Combined with a read-only root filesystem
# in the Kubernetes manifests, the container has no writable path at all.
USER nonroot:nonroot

EXPOSE 8080

ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    APP_ENV=production \
    LOG_FORMAT=json

# No shell is available, so the probe invokes the binary directly. Kubernetes
# uses its own httpGet probes and ignores this instruction.
ENTRYPOINT ["/app/server"]
