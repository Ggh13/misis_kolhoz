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

### Farmer Data

#### POST /upload_data
Загружает данные из Excel файла `internal/moked_data/farmers_sku.xlsx` в базу данных PostgreSQL.

**Response:** `200 OK` - "Successfully loaded data from excel"

#### GET /farmer_data/:id
Возвращает информацию о конкретном фермере по ID (organization_id из xlsx).

**Response:** `200 OK` - JSON с данными фермера и его продукцией

### Vector Operations (Qdrant)

#### POST /vector/:id
Создать вектор.

**Request Body:**
```json
{
  "vector": [0.1, 0.2, 0.3, ...]
}
```

**Response:** `201 Created`
```json
{"id": 1}
```

#### GET /vector/:id
Получить вектор по ID.

**Response:** `200 OK`
```json
{
  "id": 1,
  "vector": [0.1, 0.2, 0.3, ...]
}
```

**Response:** `404 Not Found` - если вектор не найден

#### PUT /vector/:id
Обновить вектор.

**Request Body:**
```json
{
  "vector": [0.1, 0.2, 0.3, ...]
}
```

**Response:** `200 OK`
```json
{"id": 1}
```

#### DELETE /vector/:id
Удалить вектор.

**Response:** `204 No Content`

#### POST /vector/search
Поиск ближайших векторов.

**Request Body:**
```json
{
  "vector": [0.1, 0.2, 0.3, ...],
  "limit": 10
}
```

**Response:** `200 OK`
```json
{"ids": [1, 5, 3, ...]}
```

#### POST /vector/distance
Вычислить расстояние между двумя векторами.

**Request Body:**
```json
{
  "vector_a": [0.1, 0.2, 0.3, ...],
  "vector_b": [0.1, 0.2, 0.3, ...]
}
```

**Response:** `200 OK`
```json
{
  "cosine": 0.95,
  "euclidean": 0.2
}
```

### Health Check

#### GET /health
Проверка здоровья сервиса.

**Response:** `200 OK`
```json
{"status": "ok"}
```

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

## Конфигурация

Настройки хранятся в `config.yaml`:

```yaml
rest_host: "0.0.0.0"
rest_port: "8080"

postgres:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "password"
  database: "mydb"

qdrant:
  host: "localhost"
  port: 6333
  collection_name: "vectors"
  vector_size: 1536
```

### Параметры Qdrant

- `host` - хост Qdrant сервера
- `port` - порт gRPC (по умолчанию 6333)
- `collection_name` - имя коллекции (по умолчанию "vectors")
- `vector_size` - размерность векторов (по умолчанию 1536)
- `distance_metric` - метрика расстояния: Cosine