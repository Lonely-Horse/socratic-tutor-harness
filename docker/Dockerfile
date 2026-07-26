FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tutor-service cmd/tutor/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/tutor-service .

COPY --from=builder /app/prompts ./prompts

RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8083

ENTRYPOINT ["./tutor-service"]