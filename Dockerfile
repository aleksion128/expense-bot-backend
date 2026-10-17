# Build frontend
FROM node:22-alpine AS web
WORKDIR /web
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Build API
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates && adduser -D -s /bin/sh appuser
WORKDIR /app
COPY --from=builder /main /app/main
COPY --from=builder /app/migrations /app/migrations
COPY --from=web /web/dist /app/frontend/dist
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8443
ENV ADDR=:8443
ENV MIGRATIONS_DIR=/app/migrations
ENV STATIC_DIR=/app/frontend/dist
ENV GIN_MODE=release
CMD ["/app/main"]
