FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM alpine:latest

# Установка ca-certificates
RUN apk --no-cache add ca-certificates

# Создание пользователя
RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/main .

# Меняем владельца
RUN chown appuser:appuser main

# Переключаемся на непривилегированного пользователя
USER appuser

# Открываем порт
EXPOSE 8443

# Переменные окружения
ENV GIN_MODE=debug

# Запуск
CMD ["./main"]