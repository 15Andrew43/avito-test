# Tender Service API

Данный проект реализует **Сервис управления тендерами**, в котором компании могут создавать тендеры на услуги, а пользователи или другие компании могут подавать свои предложения. Сервис позволяет управлять тендерами и предложениями, добавлять отзывы и выполнять откат версии как для тендеров, так и для предложений.

## Проблема

Компания Авито нуждается в сервисе, который позволяет бизнесу создавать тендеры на оказание различных услуг. Пользователи или другие компании могут предлагать свои условия для участия в тендере, предлагая конкурентные цены или услуги.

## Возможности

- **Управление тендерами**: создание, редактирование, публикация, закрытие тендера, а также откат версии.
- **Управление предложениями**: создание, редактирование, публикация, отмена предложения, согласование/отклонение.
- **Отзывы**: возможность оставлять и просматривать отзывы на предложения.
- **Откат версии**: поддержка отката версий как для тендеров, так и для предложений.

## Стек технологий

- Любой язык программирования (в проекте используется Go).
- База данных: PostgreSQL.


## Настройка через переменные окружения

Для работы приложения необходимо указать следующие переменные окружения в файле `.env`:

```dotenv
POSTGRES_HOST=rc1b-5xmqy6bq501kls4m.mdb.yandexcloud.net
POSTGRES_PORT=6432
POSTGRES_DB=cnrprod1725885797-team-78052
POSTGRES_USER=cnrprod1725885797-team-78052
POSTGRES_PASSWORD=cnrprod1725885797-team-78052
POSTGRES_TARGET_SESSION_ATTRS=read-write
SERVER_ADDRESS=0.0.0.0:8080
```

### Описание переменных:

- **POSTGRES_HOST**: Хост для подключения к базе данных PostgreSQL.
- **POSTGRES_PORT**: Порт для подключения к базе данных PostgreSQL (по умолчанию `6432`).
- **POSTGRES_DB**: Имя базы данных, которая будет использоваться приложением.
- **POSTGRES_USER**: Имя пользователя для подключения к базе данных PostgreSQL.
- **POSTGRES_PASSWORD**: Пароль для подключения к базе данных PostgreSQL.
- **POSTGRES_TARGET_SESSION_ATTRS**: Опции подключения (например, `read-write`).
- **SERVER_ADDRESS**: Адрес и порт, который будет слушать HTTP сервер (по умолчанию `0.0.0.0:8080`).



## Основные сущности

### Пользователь (User)

```sql
CREATE TABLE employee (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Организация (Organization)

```sql
CREATE TYPE organization_type AS ENUM (
    'IE', 'LLC', 'JSC'
);

CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type organization_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
);
```

### Тендер (Tender)

```sql
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tender_status') THEN
        CREATE TYPE tender_status AS ENUM ('CREATED', 'PUBLISHED', 'CLOSED');
    END IF;
END $$;

CREATE TABLE tender (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    service_type VARCHAR(50),
    status tender_status DEFAULT 'CREATED',
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    creator_id UUID REFERENCES employee(id) ON DELETE SET NULL,
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tender_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    service_type VARCHAR(50),
    status tender_status,
    organization_id UUID,
    creator_id UUID,
    version INT,
    updated_at TIMESTAMP
);

CREATE OR REPLACE FUNCTION save_tender_version()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO tender_history (tender_id, name, description, service_type, status, organization_id, creator_id, version, updated_at)
    SELECT OLD.id, OLD.name, OLD.description, OLD.service_type, OLD.status, OLD.organization_id, OLD.creator_id, OLD.version, OLD.updated_at;

    IF TG_OP = 'UPDATE' THEN
        NEW.version := OLD.version + 1;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tender_update_trigger
BEFORE UPDATE ON tender
FOR EACH ROW
EXECUTE FUNCTION save_tender_version();
```

### Предложение (Bid)

```sql
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_author_type') THEN
        CREATE TYPE bid_author_type AS ENUM ('User', 'Organization');
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'bid_status') THEN
        CREATE TYPE bid_status AS ENUM ('CREATED', 'PUBLISHED', 'CANCELED');
    END IF;
END $$;

CREATE TABLE bid (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE,
    author_type bid_author_type,
    description TEXT,
    status bid_status DEFAULT 'CREATED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Отзывы (Bid Review)

```sql
CREATE TABLE bid_review (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bid_id UUID REFERENCES bid(id) ON DELETE CASCADE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Запуск проекта

1. Сборка и запуск контейнера:

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tender-service cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /tender-service /tender-service
COPY .env .env

EXPOSE 8080

CMD ["/tender-service"]
```

2. Применение миграций перед запуском:

```bash
psql -h <host> -p <port> -U <user> -d <database> -f migrations/init.sql
```

3. Запуск сервиса:

```bash
docker build -t tender-service .
docker run -p 8080:8080 --env-file .env tender-service
```

Для тестирования всех перечисленных ручек с использованием `curl`, можно выполнить следующие запросы:

## Тесты

### 1. Проверка доступности сервера (`GET /api/ping`)

```bash
curl -X GET http://localhost:8080/api/ping
```

### 2. Получение списка тендеров (`GET /api/tenders`)

```bash
curl -X GET http://localhost:8080/api/tenders
```

### 3. Создание нового тендера (`POST /api/tenders/new`)

```bash
curl -X POST http://localhost:8080/api/tenders/new \
     -H "Content-Type: application/json" \
     -d '{
           "name": "Тендер на строительство",
           "description": "Описание тендера",
           "serviceType": "Construction",
           "status": "CREATED",
           "organizationId": "550e8400-e29b-41d4-a716-446655440021",
           "creatorUsername": "user1"
         }'
```

### 4. Получение тендеров пользователя (`GET /api/tenders/my`)

```bash
curl -X GET "http://localhost:8080/api/tenders/my?username=user1"
```

### 5. Получение статуса тендера (`GET /api/tenders/{tenderId}/status`)

```bash
curl -X GET "http://localhost:8080/api/tenders/9cd20057-0e57-42d5-a804-556267f6e4d3/status?username=user1"
```

### 6. Обновление статуса тендера (`PUT /api/tenders/{tenderId}/status`)

```bash
curl -X PUT "http://localhost:8080/api/tenders/9cd20057-0e57-42d5-a804-556267f6e4d3/status?status=PUBLISHED&username=user1"
```

### 7. Редактирование тендера (`PATCH /api/tenders/{tenderId}/edit`)

```bash
curl -X PATCH "http://localhost:8080/api/tenders/9cd20057-0e57-42d5-a804-556267f6e4d3/edit?username=user1" \
    -H "Content-Type: application/json" \
    -d '{
    "name": "Обновленное название тендера",
    "description": "Обновленное описание тендера"
    }'
```

### 8. Откат версии тендера (`POST /api/tenders/{tenderId}/rollback/{version}`)

```bash
curl -X POST "http://localhost:8080/api/tenders/9cd20057-0e57-42d5-a804-556267f6e4d3/rollback/1?username=user1"
```

### 9. Создание нового предложения (`POST /api/bids/new`)

```bash
curl -X POST http://localhost:8080/api/bids/new \
     -H "Content-Type: application/json" \
     -d '{
           "description": "Предложение на тендер",
           "tenderId": "550e8400-e29b-41d4-a716-446655440000",
           "organizationId": "550e8400-e29b-41d4-a716-446655440021",
           "userId": "550e8400-e29b-41d4-a716-446655440002",
           "authorType": "User"
         }'
```

### 10. Получение предложений пользователя (`GET /api/bids/my`)

```bash
curl -X GET "http://localhost:8080/api/bids/my?username=user1"
```

### 11. Получение списка предложений по тендеру (`GET /api/bids/{tenderId}/list`)

```bash
curl -X GET "http://localhost:8080/api/bids/9cd20057-0e57-42d5-a804-556267f6e4d3/list?username=user1"
```

### 12. Получение статуса предложения (`GET /api/bids/{bidId}/status`)

```bash
curl -X GET http://localhost:8080/api/bids/eef7c490-8e0c-4dc2-b7b5-f30a8cb94593/status?username=user2
```

### 13. Обновление статуса предложения (`PUT /api/bids/{bidId}/status`)

```bash
curl -X PUT http://localhost:8080/api/bids/eef7c490-8e0c-4dc2-b7b5-f30a8cb94593/status?username=user2 \
     -H "Content-Type: application/json" \
     -d '{
           "status": "PUBLISHED"
         }'
```

### 14. Редактирование предложения (`PATCH /api/bids/{bidId}/edit`)

```bash
curl -X PATCH http://localhost:8080/api/bids/eef7c490-8e0c-4dc2-b7b5-f30a8cb94593/edit \
     -H "Content-Type: application/json" \
     -d '{
           "description": "Обновленное описание предложения"
         }'
```

### 15. Оставление отзыва на предложение (`PUT /api/bids/{bidId}/feedback`)

```bash
curl -X PUT "http://localhost:8080/api/bids/eef7c490-8e0c-4dc2-b7b5-f30a8cb94593/feedback?username=user1&bidFeedback=Отличная работа"
```

Эти команды позволяют протестировать все доступные эндпоинты в приложении с помощью `curl`. Не забудьте заменить значения идентификаторов тендера и предложения на реальные при тестировании.