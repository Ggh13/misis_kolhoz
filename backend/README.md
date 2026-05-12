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

### Loyalty System (Client Bonuses)

#### POST /load_orders
Загружает данные заказов из Excel файла `internal/moked_data/orders.xlsx` и начисляет бонусы клиентам.
Бонус = 5% от суммы заказа, если заказ >= 1000 руб.

**Response:** `200 OK` - "Successfully loaded orders and bonuses"

#### GET /clients
Возвращает список всех клиентов.

**Response:** `200 OK`
```json
[
  {"id": 1, "name": "", "email": "", "phone": "", "created_at": "2026-05-11T20:18:13Z"},
  ...
]
```

#### GET /client_bonus/:id
Возвращает информацию о бонусах клиента по ID.

**Response:** `200 OK`
```json
{
  "client": {"id": 2, "name": "", "email": "", "phone": "", "created_at": "..."},
  "balance": 3340,
  "transactions": [
    {"id": 19, "client_id": 2, "type": "accrual", "amount": 480, "order_id": 0, "created_at": "..."},
    ...
  ]
}
```

**Response:** `404 Not Found` - если клиент не найден

#### POST /spend_bonus/:id
Списать бонусы клиента.

**Request Body:**
```json
{
  "amount": 100
}
```

**Response:** `200 OK`
```json
{"message": "bonus spent successfully"}
```

**Response:** `400 Bad Request` - недостаточно бонусов или ошибка

### Recommendation System

#### GET /recommendations/:id
Возвращает сгенерированные рекомендации для клиента по его `User_ID` из `orders.xlsx`.

Сервис работает в фоне:
- при старте backend делает первую попытку расчёта рекомендаций;
- затем пересчитывает рекомендации периодически (раз в минуту);
- endpoint читает уже готовый кэш в памяти (а не считает рекомендации на каждый запрос).

Источник данных:
- покупки: `internal/moked_data/orders.xlsx`;
- каталог товаров: таблица `farmer_products` (заполняется через `POST /upload_data` из `farmers_sku.xlsx`).

Логика рекомендаций (текущая версия):
- берётся последняя покупка клиента по категории;
- если она была примерно месяц назад (окно 28..45 дней),
  подбирается товар из той же категории, который клиент ещё не покупал.

**Важно:** перед проверкой рекомендаций нужно вызвать `POST /upload_data`, иначе каталог товаров в БД может быть пустым.

**Response:** `200 OK`
```json
{
  "client_id": 3,
  "recommendations": [
    {
      "product_id": 81,
      "farmer_id": 183,
      "product_name": "Саган-дайля сушеная",
      "category": "Бакалея",
      "unit": "шт",
      "price": 183,
      "quantity": 0
    }
  ]
}
```

**Response:** `404 Not Found` - клиент не найден в текущем кэше рекомендаций

**Response:** `503 Service Unavailable` - рекомендации ещё не готовы (например, до первого успешного фонового пересчёта)

Пример ручной проверки:
```bash
curl -X POST http://localhost:8080/upload_data
sleep 70
curl http://localhost:8080/recommendations/1
curl http://localhost:8080/recommendations/3
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