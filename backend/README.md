# Go Project Template

## Структура проекта

```
cmd/                    # Точка входа
internal/config/        # Конфигурация
internal/transport/rest/ # HTTP роутер
pkg/logger/             # Логгер
pkg/postgres/           # PostgreSQL клиент
pkg/mongo/              # MongoDB клиент
pkg/s3Storage/          # S3 клиент
config.yaml             # Конфигурация
Dockerfile              # Сборка образа
docker-compose.yml      # Запуск сервисов
Makefile                # Команды
```

## Требования

- Go 1.22+
- Docker + Docker Compose

## Запуск

### Локально

```bash
# Установка зависимостей
go mod download

# Запуск
go run ./cmd
```

### Docker

```bash
# Сборка и запуск
make up

# Остановка
make down

# Пересборка
make build
```

## Команды Makefile

| Команда | Описание |
|---------|----------|
| `make build` | Сборка Docker образа |
| `make up` | Запуск всех сервисов |
| `make down` | Остановка всех сервисов |
| `make clean` | Остановка с удалением томов |
| `make run` | Запуск локально без Docker |
| `make test` | Запуск тестов |
| `make tidy` | Обновление зависимостей |

## Конфигурация

Настройки хранятся в `config.yaml`:

- `rest_host` / `rest_port` - адрес HTTP сервера
- `postgres` - подключение к PostgreSQL
- `mongo` - подключение к MongoDB
- `s3` - настройки S3

## Эндпоинты

- `GET /health` - проверка здоровья сервиса

## Переменные окружения

- `CONFIG_PATH` - путь к конфигурационному файлу (по умолчанию `config.yaml`)
