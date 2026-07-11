# Stage 1: Build static assets (Tailwind compilation & JS minification)
FROM node:22-alpine AS asset-builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY web/ ./web/
RUN npm run build

# Stage 2: Build the Go application
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy compiled assets from asset-builder to make sure they are embedded
COPY --from=asset-builder /app/web/assets/css/theme.css ./web/assets/css/theme.css
COPY --from=asset-builder /app/web/assets/js/app.min.js ./web/assets/js/app.min.js
COPY --from=asset-builder /app/web/assets/js/admin.min.js ./web/assets/js/admin.min.js
COPY --from=asset-builder /app/web/assets/js/components/stepper.min.js ./web/assets/js/components/stepper.min.js
COPY --from=asset-builder /app/web/assets/js/components/contact.min.js ./web/assets/js/components/contact.min.js

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api/main.go

# Stage 3: Minimal runtime container
FROM alpine:3.24.1

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main ./main
COPY --from=builder /app/web ./web
COPY --from=builder /app/locales ./locales

ENV PORT=8080

USER nobody:nogroup

EXPOSE 8080

ENTRYPOINT ["./main"]
