# misis_kolhoz

## Запуск

### Docker Compose (рекомендуется)

```bash
docker-compose up --build
```

Приложение будет доступно по адресу: http://localhost:8080

### Запуск без Docker

```bash
cd backend
go mod download
go run cmd/main.go
```

## Остановка

```bash
docker-compose down
```

Для остановки с удалением контейнеров (без volumes):
```bash
docker-compose down -v
```

## Переменные окружения

Настройки хранятся в `backend/config.yaml`. Основные параметры:

- `rest_host` - хост REST API (по умолчанию 0.0.0.0)
- `rest_port` - порт REST API (по умолчанию 8080)
- PostgreSQL, MongoDB, S3 настраиваются в соответствующих секциях