# Stage 1: Asset Builder (Bun 1-alpine)
FROM oven/bun:1-alpine AS asset-builder
WORKDIR /app
COPY package.json bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache bun install --frozen-lockfile
COPY web/ ./web/
RUN bun run build

# Stage 2: Go App Builder (Go 1.26-alpine)
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
COPY --from=asset-builder /app/web/assets/dist ./web/assets/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -extldflags '-static'" -o api ./cmd/api/main.go

# Stage 3: Immutable Scratch Runtime
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/api ./api
ENV PORT=8080
USER appuser
EXPOSE 8080
ENTRYPOINT ["./api"]
