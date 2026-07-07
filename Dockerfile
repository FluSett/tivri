FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/api/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main ./main
COPY --from=builder /app/web ./web
COPY --from=builder /app/locales ./locales

ENV PORT=8080

USER nobody:nogroup

EXPOSE 8080

ENTRYPOINT ["./main"]
