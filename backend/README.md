# misis_kolhoz Backend

## Запуск

### Docker Compose (рекомендуется)

```bash
docker-compose up --build -d
```

Фронтенд: http://localhost:80
API: http://localhost:8080

### Без Docker

```bash
cd backend
go mod download
go run cmd/main.go
```

## Структура проекта

```
backend/
├── cmd/main.go                 — точка входа
├── config.yaml                 — конфигурация
├── Dockerfile                  — сборка Docker-образа
├── go.mod / go.sum
├── internal/
│   ├── config/                 — загрузка конфигурации
│   ├── events/                 — модуль ивентов/инфоповодов
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   ├── farmer/                 — модуль фермеров и продукции
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   ├── loyalty/                — модуль лояльности (клиенты, бонусы)
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   ├── recommendation/         — рекомендации товаров
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   ├── vector/                 — векторные операции (pgvector)
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   └── transport/rest/         — REST-роутер (Gin)
├── moked_data/
│   ├── events.xlsx             — календарь ивентов/инфоповодов
│   ├── farmers_sku.xlsx        — данные фермеров и продукции
│   └── orders.xlsx             — данные заказов
└── pkg/
    ├── logger/
    ├── postgres/
    └── qdrant/                  — (устарело, оставлено для совместимости)
```

## API Endpoints

---

### Health Check

#### `GET /health`

Проверка работоспособности сервиса.

**Request:**
```
GET http://localhost:8080/health
```

**Response:** `200 OK`
```json
{
  "status": "ok"
}
```

**cURL:**
```bash
curl http://localhost:8080/health
```

---

### Farmer Data

#### `POST /upload_data`

Загружает данные из Excel файла `moked_data/farmers_sku.xlsx` в PostgreSQL. Создаёт таблицы `farmers` и `farmer_products` (если не существуют), сохраняет `farmer_description`/`product_description`, а таблица `product_embeddings` заполняется нулевыми векторами размерности 384.

**Request:**
```
POST http://localhost:8080/upload_data
```

**Response:** `200 OK`
```json
"Successfully loaded data from excel"
```

**cURL:**
```bash
curl -X POST http://localhost:8080/upload_data
```

---

### Events / Info Occasions

#### `POST /upload_events`

Загружает данные из Excel файла `internal/moked_data/events.xlsx` в таблицу `event_embeddings`.

Схема таблицы `event_embeddings`:

| Колонка      | Тип         | Описание                              |
|--------------|-------------|---------------------------------------|
| id           | SERIAL      | Первичный ключ                        |
| event_date   | TEXT        | Дата/дата-диапазон события (нормализуется при загрузке) |
| holiday_info | TEXT        | Праздник / инфоповод                  |
| category     | TEXT        | Категория                             |
| about        | TEXT        | О чём событие                         |
| food_customs | TEXT        | Еда / обычаи                          |
| embedding    | vector(384) | Векторное представление события       |

Колонки, ожидаемые в XLSX:
- `Дата`
- `Праздник / инфоповод`
- `Категория`
- `О чём он`
- `Еда / обычаи`

Для каждой строки создаётся запись с вектором `embedding vector(384)` (по умолчанию нулевой вектор).

**Request:**
```
POST http://localhost:8080/upload_events
```

**Response:** `200 OK`
```json
"Successfully loaded events data from excel"
```

**cURL:**
```bash
curl -X POST http://localhost:8080/upload_events
```

---

#### `GET /events`

Возвращает все ивенты, отсортированные по дате по возрастанию.

**Request:**
```
GET http://localhost:8080/events
```

---

#### `GET /events/month/:year/:month`

Возвращает ивенты за конкретный месяц.

**Request:**
```
GET http://localhost:8080/events/month/2026/12
```

---

#### `GET /events/upcoming?days=30&limit=100`

Возвращает предстоящие ивенты в окне от сегодняшней даты до `today + days`.

Параметры:
- `days` (опционально, по умолчанию `30`)
- `limit` (опционально, по умолчанию `100`)

**Request:**
```
GET http://localhost:8080/events/upcoming?days=45&limit=50
```

---

#### `GET /events/category/:category`

Возвращает ивенты по категории (поиск без учёта регистра).

**Request:**
```
GET http://localhost:8080/events/category/фрукты
```

---

#### `GET /farmer_data/:id`

Возвращает информацию о конкретном фермере по ID (`organization_id` из xlsx), включая список его продукции.

**Request:**
```
GET http://localhost:8080/farmer_data/{id}
```

**Response:** `200 OK`
```json
{
  "farmer": {
    "id": 1001,
    "name": "Иванов Иван",
    "region": "Московская область",
    "address": "ул. Ленина, д. 10",
    "phone": "+7-999-123-45-67",
    "email": "ivanov@example.com"
  },
  "products": [
    {
      "id": 1,
      "farmer_id": 1001,
      "product_name": "Яблоки",
      "category": "Фрукты",
      "unit": "шт",
      "price": 150.00,
      "quantity": 0
    },
    {
      "id": 2,
      "farmer_id": 1001,
      "product_name": "Мёд",
      "category": "Сладости",
      "unit": "шт",
      "price": 800.00,
      "quantity": 0
    }
  ]
}
```

**Response (not found):** `404 Not Found`
```
"Farmer not found"
```

**cURL:**
```bash
curl http://localhost:8080/farmer_data/1001
```

---

### Vector Operations (pgvector)

Все операции с векторами работают через расширение **pgvector** в PostgreSQL (вместо ранее используемого Qdrant).

Таблица `product_embeddings` имеет структуру:

| Колонка       | Тип              | Описание                           |
|---------------|------------------|------------------------------------|
| id            | SERIAL           | Первичный ключ                     |
| product_id    | INTEGER (UNIQUE) | Ссылка на farmer_products(id)      |
| farmer_id     | INTEGER          | ID фермера                         |
| product_name  | VARCHAR(255)     | Название продукта                  |
| category      | VARCHAR(100)     | Категория продукта                 |
| unit          | VARCHAR(50)      | Единица измерения                  |
| price         | DECIMAL(10,2)    | Цена                               |
| quantity      | INTEGER          | Количество                         |
| farmer_description | TEXT       | Описание фермера из farmers_sku    |
| product_description | TEXT      | Описание продукта из farmers_sku   |
| embedding     | vector(384)      | Вектор эмбеддинга (384 измерения)  |

Поиск работает через косинусное расстояние с использованием **IVFFlat-индекса**.

---

#### `POST /vector/:id`

Создать или обновить (upsert) вектор для продукта по `product_id`. Автоматически подтягивает все данные продукта из таблицы `farmer_products`.

**Request:**
```
POST http://localhost:8080/vector/{product_id}
Content-Type: application/json
```
```json
{
  "product_id": 42,
  "embedding": [0.1, 0.2, 0.3, ..., 0.05]
}
```
*(массив ровно из 384 чисел float32)*

**Response:** `201 Created`
```json
{
  "product_id": 42
}
```

**Response (ошибка):** `400 Bad Request`
```json
{
  "error": "embedding size must be 384"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/vector/42 \
  -H "Content-Type: application/json" \
  -d '{"product_id": 42, "embedding": [0.1, 0.2, 0.3, ...]}'
```

---

#### `GET /vector/:id`

Получить данные эмбеддинга для продукта по `product_id`.

**Request:**
```
GET http://localhost:8080/vector/{product_id}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "product_id": 42,
  "farmer_id": 1001,
  "product_name": "Яблоки",
  "category": "Фрукты",
  "unit": "шт",
  "price": 150.00,
  "quantity": 50,
  "embedding": [0.1, 0.2, 0.3, ..., 0.05]
}
```

**Response (not found):** `404 Not Found`
```json
{
  "error": "vector not found"
}
```

**cURL:**
```bash
curl http://localhost:8080/vector/42
```

---

#### `PUT /vector/:id`

Обновить вектор для продукта. Полностью перезаписывает embedding.

**Request:**
```
PUT http://localhost:8080/vector/{product_id}
Content-Type: application/json
```
```json
{
  "embedding": [0.5, 0.1, 0.8, ..., 0.02]
}
```

**Response:** `200 OK`
```json
{
  "product_id": 42
}
```

**cURL:**
```bash
curl -X PUT http://localhost:8080/vector/42 \
  -H "Content-Type: application/json" \
  -d '{"embedding": [0.5, 0.1, 0.8, ...]}'
```

---

#### `DELETE /vector/:id`

Удалить запись из `product_embeddings` по `product_id`.

**Request:**
```
DELETE http://localhost:8080/vector/{product_id}
```

**Response:** `204 No Content`

**cURL:**
```bash
curl -X DELETE http://localhost:8080/vector/42
```

---

#### `POST /vector/search`

Поиск ближайших продуктов по косинусному расстоянию. Возвращает список продуктов, отсортированных по возрастанию расстояния (меньше значение = ближе по смыслу).

**Request:**
```
POST http://localhost:8080/vector/search
Content-Type: application/json
```
```json
{
  "vector": [0.1, 0.2, 0.3, ..., 0.05],
  "limit": 10
}
```

**Response:** `200 OK`
```json
{
  "products": [
    {
      "id": 1,
      "product_id": 42,
      "farmer_id": 1001,
      "product_name": "Яблоки",
      "category": "Фрукты",
      "unit": "кг",
      "price": 150.00,
      "quantity": 50
    },
    {
      "id": 2,
      "product_id": 56,
      "farmer_id": 1002,
      "product_name": "Груши",
      "category": "Фрукты",
      "unit": "кг",
      "price": 200.00,
      "quantity": 30
    }
  ]
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/vector/search \
  -H "Content-Type: application/json" \
  -d '{"vector": [0.1, 0.2, ...], "limit": 10}'
```

---

#### `POST /vector/distance`

Вычислить косинусное и евклидово расстояние между двумя векторами.

**Request:**
```
POST http://localhost:8080/vector/distance
Content-Type: application/json
```
```json
{
  "vector_a": [0.1, 0.2, 0.3],
  "vector_b": [0.4, 0.5, 0.6]
}
```

**Response:** `200 OK`
```json
{
  "cosine": 0.974,
  "euclidean": 0.520
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/vector/distance \
  -H "Content-Type: application/json" \
  -d '{"vector_a": [0.1, 0.2, 0.3], "vector_b": [0.4, 0.5, 0.6]}'
```

---

### Loyalty System (Client Bonuses)

#### `POST /load_orders`

Загружает данные заказов из Excel файла `moked_data/orders.xlsx` и начисляет бонусы клиентам. Бонус = 5% от суммы заказа, если сумма ≥ 1000 ₽.

**Request:**
```
POST http://localhost:8080/load_orders
```

**Response:** `200 OK`
```json
"Successfully loaded orders and bonuses"
```

**cURL:**
```bash
curl -X POST http://localhost:8080/load_orders
```

---

#### `GET /clients`

Возвращает список всех клиентов, зарегистрированных в системе.

**Request:**
```
GET http://localhost:8080/clients
```

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Алексей Петров",
    "email": "alex@example.com",
    "phone": "+7-900-111-22-33",
    "created_at": "2026-05-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "Мария Сидорова",
    "email": "maria@example.com",
    "phone": "+7-900-444-55-66",
    "created_at": "2026-05-11T08:30:00Z"
  }
]
```

**cURL:**
```bash
curl http://localhost:8080/clients
```

---

#### `GET /client_bonus/:id`

Возвращает информацию о бонусах клиента по ID, включая текущий баланс и историю транзакций.

**Request:**
```
GET http://localhost:8080/client_bonus/{id}
```

**Response:** `200 OK`
```json
{
  "client": {
    "id": 1,
    "name": "Алексей Петров",
    "email": "alex@example.com",
    "phone": "+7-900-111-22-33",
    "created_at": "2026-05-10T12:00:00Z"
  },
  "balance": 3340,
  "transactions": [
    {
      "id": 19,
      "client_id": 1,
      "type": "accrual",
      "amount": 480,
      "order_id": 100,
      "created_at": "2026-05-11T14:22:00Z"
    },
    {
      "id": 20,
      "client_id": 1,
      "type": "debit",
      "amount": 200,
      "order_id": 0,
      "created_at": "2026-05-12T10:15:00Z"
    }
  ]
}
```

**Response (not found):** `404 Not Found`
```json
{
  "error": "client not found"
}
```

**cURL:**
```bash
curl http://localhost:8080/client_bonus/1
```

---

#### `POST /spend_bonus/:id`

Списать бонусы с баланса клиента.

**Request:**
```
POST http://localhost:8080/spend_bonus/{id}
Content-Type: application/json
```
```json
{
  "amount": 500
}
```

**Response:** `200 OK`
```json
{
  "message": "bonus spent successfully"
}
```

**Response (ошибка):** `400 Bad Request`
```json
{
  "error": "insufficient bonus balance"
}
```

**cURL:**
```bash
curl -X POST http://localhost:8080/spend_bonus/1 \
  -H "Content-Type: application/json" \
  -d '{"amount": 500}'
```

---

### Recommendation System

#### `GET /recommendations/:id`

Получить персональные рекомендации товаров для клиента на основе его истории покупок. Рекомендации основаны на категориях недавних покупок (окно 28–45 дней). Для каждой подходящей категории выбирается один товар из каталога, который клиент ещё не покупал.

**Важно:** перед проверкой рекомендаций нужно вызвать `POST /upload_data`, чтобы каталог товаров `farmer_products` был заполнен.

Сервис работает с кэшем:
- При старте backend делает первую попытку расчёта рекомендаций
- Затем пересчитывает автоматически раз в минуту
- Endpoint читает готовый кэш из памяти (не считает на каждый запрос)

**Request:**
```
GET http://localhost:8080/recommendations/{client_id}
```

**Response:** `200 OK`
```json
{
  "client_id": 1,
  "recommendations": [
    {
      "product_id": 56,
      "farmer_id": 1002,
      "product_name": "Груши",
      "category": "Фрукты",
      "unit": "кг",
      "price": 200.00,
      "quantity": 30
    },
    {
      "product_id": 78,
      "farmer_id": 1003,
      "product_name": "Минеральная вода",
      "category": "Напитки",
      "unit": "шт",
      "price": 80.00,
      "quantity": 100
    }
  ]
}
```

**Response (не найден):** `404 Not Found`
```json
{
  "error": "client not found"
}
```

**Response (рекомендации ещё не посчитаны):** `503 Service Unavailable`
```json
{
  "error": "recommendations not ready"
}
```

**cURL:**
```bash
# Сначала загрузить данные
curl -X POST http://localhost:8080/upload_data

# Подождать ~70 секунд для расчёта рекомендаций
sleep 70

# Получить рекомендации для клиента 1 и 3
curl http://localhost:8080/recommendations/1
curl http://localhost:8080/recommendations/3
```

---

## Быстрые команды

```bash
# Запуск всех сервисов
docker-compose up --build -d

# Остановка
docker-compose down

# Остановка с удалением данных
docker-compose down -v

# Логи
docker-compose logs -f app
docker-compose logs -f frontend

# Сборка без кэша
docker-compose up --build --force-recreate -d

# Локальная разработка (без Docker)
cd backend && go run cmd/main.go

# Тесты
go test ./... -v
go test ./internal/farmer/... -v
```

## Конфигурация

Файл `config.yaml`:

```yaml
rest_host: "0.0.0.0"
rest_port: "8080"

postgres:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "password"
  database: "mydb"
  min_conns: 2
  max_conns: 10
```

## Стек технологий

- **Go 1.24** — backend
- **Gin** — HTTP-фреймворк
- **pgx** — PostgreSQL-драйвер
- **pgvector** — расширение PostgreSQL для векторных операций
- **Excelize** — чтение Excel-файлов
- **React 19 + Vite** — фронтенд
- **Docker Compose** — оркестрация контейнеров
- **Nginx** — фронтенд-сервер в Docker
