# misis_kolhoz Backend

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

## API Endpoints

### POST /upload_data
Загружает данные из Excel файла `internal/moked_data/farmers_sku.xlsx` в базу данных PostgreSQL.

**Response:** `200 OK` - "Successfully loaded data from excel"

### GET /farmer_data/:id
Возвращает информацию о конкретном фермере по ID (organization_id из xlsx).

**Response:** `200 OK` - JSON с данными фермера и его продукцией

## Остановка

```bash
docker-compose down
```

Для остановки с удалением контейнеров (без volumes):
```bash
docker-compose down -v
```

## Тесты

Запуск всех тестов:
```bash
go test ./... -v
```

Запуск тестов конкретного пакета:
```bash
go test ./internal/farmer/handler/... -v
```

## Переменные окружения

Настройки хранятся в `config.yaml`. Основные параметры:

- `rest_host` - хост REST API (по умолчанию 0.0.0.0)
- `rest_port` - порт REST API (по умолчанию 8080)
- PostgreSQL, MongoDB, S3 настраиваются в соответствующих секциях