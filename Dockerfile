# Stage 1: Build static assets (Tailwind compilation & JS minification)
FROM node:22-alpine AS asset-builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY web/ ./web/
RUN npm run build

# Stage 2: Build the Go application
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata
# Create a non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy compiled assets from asset-builder to make sure they are embedded
COPY --from=asset-builder /app/web/assets ./web/assets

# Build with -trimpath and remove buildid for minimal size
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -buildid=" -o main ./cmd/api/main.go

# Stage 3: Minimal runtime container (scratch)
FROM scratch

# Copy certificates, tzdata, and user from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

WORKDIR /app

COPY --from=builder /app/main ./main

ENV PORT=8080

USER appuser

EXPOSE 8080

ENTRYPOINT ["./main"]
