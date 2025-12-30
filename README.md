# Geomessanging Service

Сервис для управления геолокационными инцидентами и проверки координат пользователей на попадание в опасные зоны.

## Описание
- Создания и управления инцидентами (опасными зонами) с географическими координатами
- Проверки координат пользователей на попадание в опасные зоны
- Получения статистики по зонам
- Асинхронной отправки webhook-уведомлений при обнаружении пользователей в опасных зонах

## Технологии

- **Go 1.25**
- **PostgreSQL 15** с расширением PostGIS для работы с геоданными
- **Redis 7** 
- **Docker** и **Docker Compose** 

## Установка и запуск

### Запуск через Docker Compose

1. Создайте файл `.env` командой `cp .env-test .env`
2. Запустите все сервисы `make docker-up` или `docker compose up --build -d`
3. Остановка сервисов `make docker-down` или `docker compose down -v`


## Миграции

Миграции базы данных выполняются автоматически при запуске приложения с использованием [goose](https://github.com/pressly/goose).
Миграции находятся в директории `migrations/`

## Настройка Cloudpub
**Вместо ngrok я использовал cloudpub, но это не ограничивает вас в использование ngrok. Можно использовать оба варианта**

Инструкции по установке Cloudpub находятся в файле [настройка](setup-cloudpub.md).

Cloudpub предоставит публичный URL (например, `https://xxxxx.cloudpub.io`). Обновите переменную `WEBHOOK_URL` в `.env` файле, если необходимо использовать внешний URL.

## API Endpoints

Сервис: `http://localhost:8080`

Вебхук сервис `http://localhost:9090`

Swagger UI (Документация API) `http:/localhost:8082`

### Аутентификация

Endpoints, отвечающие за управление инцидентами, требуют заголовок `X-API-Key` с валидным API ключом.

### 1. Health Check

Проверка работоспособности сервиса.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/system/health
```

**Response:**
```
200 OK
```

### 2. Создание инцидента

Создание новой опасной зоны.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key" \
  -d '{
    "title": "Атака демогоргонов",
    "description": "Большое скопление монстров около парка",
    "lat": 41.2192,
    "long": 86.4892,
    "radius_m": 500,
    "active": true
  }'
```

**Response:**
```json
{
  "Incedent": {
    "id": 1,
    "title": "Атака демогоргонов",
    "description": "Большое скопление монстров около парка",
    "lat": 41.2192,
    "long": 86.4892,
    "radius_m": 500,
    "active": true,
    "created_at": "1983-11-16T10:30:00Z",
    "updated_at": "1983-11-16T10:30:00Z"
  }
}
```

### 3. Получение инцидента по ID

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/incidents/1 \
  -H "X-API-Key: your-secret-api-key"
```

**Response:**
```json
{
  "Incedent": {
    "id": 1,
    "title": "Атака демогоргонов",
    "description": "Большое скопление монстров около парка",
    "lat": 41.2192,
    "long": 86.4892,
    "radius_m": 500,
    "active": true,
    "created_at": "1983-11-16T10:30:00Z",
    "updated_at": "1983-11-16T10:30:00Z"
  }
}
```

### 4. Получение списка инцидентов (с пагинацией)

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/incidents?page=1&limit=10" \
  -H "X-API-Key: your-secret-api-key"
```

**Response:**
```json
{
  "incidents": [
    {
      "id": 1,
      "title": "Атака демогоргонов",
      "description": "Большое скопление монстров около парка",
      "lat": 41.2192,
      "long": 86.4892,
      "radius_m": 500,
      "active": true,
      "created_at": "1983-11-16T10:30:00Z",
      "updated_at": "1983-11-16T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

### 5. Обновление инцидента

**Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/incidents/1 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key" \
  -d '{
    "title": "Атака демогоргонов (убили)",
    "description": "Вилл Байрес с нетрадиционной ориентацией зачистил зону",
    "lat": 41.2192,
    "long": 86.4892,
    "radius_m": 500,
    "active": false
  }'
```

**Response:**
```json
{
  "Incedent": {
    "id": 1,
    "title": "Атака демогоргонов (убили)",
    "description": "Вилл Байрес с нетрадиционной ориентацией зачистил зону",
    "lat": 41.2192,
    "long": 86.4892,
    "radius_m": 500,
    "active": false,
    "created_at": "1983-11-16T10:30:00Z",
    "updated_at": "1983-1-15T11:00:00Z"
  }
}
```

### 6. Удаление инцидента

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/incidents/1 \
  -H "X-API-Key: your-secret-api-key"
```

**Response:**
```
200 OK
```

### 7. Проверка координат

Проверяет, находится ли пользователь в опасной зоне. Если да, отправляет webhook-уведомление.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/location/check \
  -H "Content-Type: application/json" \
  -d '{
    "used_id": "Lucas",
    "lat": 41.2192,
    "long": 86.491
  }'
```

**Response:**
```json
{
  "id": 1,
  "user_id": "Lucas",
  "checked_at": "2025-12-30T21:54:22.564565Z",
  "lat": 41.2192,
  "long": 86.491,
  "in_danger_zone": true,
  "nearest_id": 2
}
```

### 8. Статистика по зонам

Получает статистику по количеству пользователей в каждой зоне за указанный временной период.

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/incidents/stats
```

**Response:**
```json
{
  "Stats": [
    {
      "zone_id": 1,
      "user_count": 1
    },
  ]
}
```

## Webhook

При проверке координат, если пользователь находится в опасной зоне, система асинхронно отправляет webhook-уведомление на указанный URL.

### Формат webhook-уведомления

**URL:** `POST {WEBHOOK_URL}`

**Body:**
```json
{
  "id": 1,
  "user_id": "Lucas",
  "lat": 41.2192,
  "long": 86.491,
  "in_danger_zone": true,
  "nearest_id": 1,
  "checked_at": "1983-1-15T11:00:00Z"
}
```

## Структура проекта

```
.
├── cmd/
│   ├── app/          # Основное приложение
│   └── webhook/      # Webhook-сервер для тестирования
│
├── docs/             # Swagger
│
├── internal/
│   ├── config/       # Конфигурация
│   ├── domain/       # Доменные модели
│   ├── handler/      # HTTP handlers
│   ├── repository/   # Репозитории для работы с БД и Redis
│   ├── service/      # Бизнес-логика
│   └── workers/      # Фоновые воркеры (webhook worker)
├── migrations/       # Миграции базы данных
├── docker-compose.yaml
├── Dockerfile
├── Dockerfile.webhook
├── Makefile
├── README.md
└── setup-cloudpub.md
```

## Разработка

### Запуск тестов

1. Запуск юнит тестов - `make test-unit`
2. Запуск интгерационных тестов - `make test-integration`
3. Запуск всех тестов - `make test`