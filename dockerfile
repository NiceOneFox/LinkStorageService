# Этап 1: Сборка (builder)
FROM golang:1.26-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарный файл
# CGO_ENABLED=0 - статическая сборка (без зависимостей на C)
# GOOS=linux - целевая операционная система
# -a - пересобрать все пакеты
# -installsuffix cgo - отключает CGO
# -o linkstorage - имя выходного файла
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o linkstorage ./cmd/server/main.go

# Этап 2: Финальный образ (run)
FROM alpine:latest

# Устанавливаем CA-сертификаты (нужны для HTTPS запросов)
# И временные зоны (для корректного времени)
RUN apk --no-cache add ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/linkstorage .

# Открываем порт (только документация, реальный проброс в docker-compose)
EXPOSE 8080

# Запускаем приложение
CMD ["./linkstorage"]