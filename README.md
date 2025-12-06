# Kursovaya AKSP — микросервисное приложение для бронирования

Прототип клиент-серверной системы бронирования спортивных залов и площадок. Архитектура построена на независимых Go-сервисах с четкими зонами ответственности, REST-взаимодействием и PostgreSQL в ключевых стораджах (identity/facility/booking).

## Сервисы и порты

| Сервис | Порт | Назначение |
| --- | --- | --- |
| API Gateway | `:8080` | Единая точка входа, проксирование запросов на внутренние сервисы |
| Identity | `:8081` | Регистрация, логин, управление токенами и профилями |
| Facility | `:8082` | Каталог площадок, расписание доступности |
| Booking | `:8083` | Бронирования, проверка пересечений, поиски доступных слотов |
| Notification | `:8084` | Логирование отправленных уведомлений (email/SMS/push) |

## Структура репозитория

```
go.work
services/
  api-gateway/
  identity/
  facility/
  booking/
  notification/
```

Каждый сервис — отдельный Go-модуль, поэтому проект управляется через `go.work`.

## Архитектура сервисов

Внутри модулей выделены четкие слои, чтобы проще эволюционировать доменную логику и адаптеры:

- `internal/app` — HTTP-серверы и хендлеры на chi, работающие только с DTO.
- `internal/dto` — структуры запросов/ответов, используемые REST-контроллерами.
- `internal/models` — доменные сущности, которыми оперируют бизнес-правила и стор.
- `internal/store` — доступ к PostgreSQL через [`github.com/Masterminds/squirrel`](https://github.com/Masterminds/squirrel), что избавляет от ручной конкатенации SQL и упрощает добавление фильтров.

Такой разрез позволяет переиспользовать модели в gRPC/GraphQL слоях и держать конвертацию данных в одном месте.

## gRPC каналы

Внутренние сервисы поднимают собственные gRPC-серверы (по умолчанию `:9081` для Identity и `:9082` для Facility). Booking сервис использует их для валидации пользователей и площадок перед созданием брони.

Параметры можно переопределить переменными окружения:

- `IDENTITY_GRPC_ADDR` — адрес gRPC сервера identity.
- `FACILITY_GRPC_ADDR` — адрес gRPC сервера facility.
- Эти же переменные читает Booking при создании gRPC-клиентов. В Docker Compose по умолчанию используются `identity:9081` и `facility:9082`.

## Быстрый старт

1. Поднимите PostgreSQL (локально или через Docker) и создайте базу `sporthub`:
   ```powershell
   docker run --name sporthub-pg -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15
   docker exec -it sporthub-pg createdb -U postgres sporthub
   ```
2. Примените миграции (можно через `psql`):
   ```powershell
   psql $Env:POSTGRES_URL -f services/identity/migrations/001_init.sql
   psql $Env:POSTGRES_URL -f services/facility/migrations/001_init.sql
   psql $Env:POSTGRES_URL -f services/booking/migrations/001_init.sql
   ```
   В качестве `$Env:POSTGRES_URL` можно использовать `postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable`.
3. Задайте переменные окружения (или полагайтесь на дефолтные DSN `postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable`):
   ```powershell
   $Env:IDENTITY_DATABASE_URL="postgres://postgres:postgres@localhost:5432/sporthub?sslmode=disable"
   $Env:FACILITY_DATABASE_URL=$Env:IDENTITY_DATABASE_URL
   $Env:BOOKING_DATABASE_URL=$Env:IDENTITY_DATABASE_URL
   ```
4. Запустите каждую службу в своем терминале:
   ```powershell
   cd services/identity; go run ./cmd/identity
   cd services/facility; go run ./cmd/facility
   cd services/booking; go run ./cmd/booking
   cd services/notification; go run ./cmd/notification
   cd services/api-gateway; go run ./cmd/api-gateway
   ```
5. При необходимости переопределите сетевые параметры:
   - `GATEWAY_ADDR`, `IDENTITY_URL`, `FACILITY_URL`, `BOOKING_URL`, `NOTIFICATION_URL`
   - `IDENTITY_DATABASE_URL`, `FACILITY_DATABASE_URL`, `BOOKING_DATABASE_URL`
   - `BOOKING_ADDR`, `IDENTITY_GRPC_ADDR`, `FACILITY_GRPC_ADDR`, `NOTIFICATION_ADDR`, и т.д.

### Запуск через Docker Compose

```
make compose-up
```

Команда собирает образы, поднимает PostgreSQL 15 (`sporthub-postgres`) и все службы в одной сети. База живёт в томе `postgres_data`, gRPC-порты (`9081`, `9082`) проброшены наружу для отладки, а сервисы получают DSN вида `postgres://postgres:postgres@postgres:5432/sporthub?sslmode=disable`. Остановить окружение и очистить данные можно командой `make compose-down`.

После старта можно накатывать миграции через штатный `psql`, например:

```powershell
make compose-migrate
```

Команда под капотом пробрасывает SQL-файлы в контейнер `postgres` через stdin. При желании можно выполнить вручную:

```powershell
docker compose exec -T postgres psql -U postgres -d sporthub < services/identity/migrations/001_init.sql
docker compose exec -T postgres psql -U postgres -d sporthub < services/facility/migrations/001_init.sql
docker compose exec -T postgres psql -U postgres -d sporthub < services/booking/migrations/001_init.sql
```

### Полезные make-команды

- `make proto` — прогоняет `protoc` для всех `.proto` файлов и обновляет gRPC-стабы.
- `make compose-migrate` — применяет SQL миграции `identity`, `facility` и `booking` внутри docker compose окружения.

## Основные REST-ручки

### API Gateway (`:8080`)
- `POST /auth/register`, `/auth/login`, `/auth/logout`, `/auth/refresh`
- `GET /users/me`, `PATCH /users/me`
- `/facilities` — полный CRUD и управление доступностью (проксируется на Facility)
- `/bookings` — создание, подтверждение, отмена (проксируется на Booking)
- `/availability/search` — проверка свободных слотов
- `/notifications` — просмотр и тестовые отправки
- `/graphql` — агрегирующий GraphQL слой (GraphiQL включён). Теперь поддерживает и чтение, и основные мутации (`register`, `login`, `updateProfile`, `createFacility`, `updateFacilityAvailability`, `createBooking`, `confirmBooking`, `cancelBooking`, `deleteBooking`). Примеры:
   ```graphql
   query FacilitiesWithBookings($city: String!) {
      facilities(city: $city) {
         id
         name
      }
      availability(facilityId: "fcl-demo") {
         start
         end
      }
      bookings(facilityId: "fcl-demo") {
         id
         status
         startsAt
      }
   }
   ```
   ```graphql
   mutation BookFacility($input: BookingInput!) {
      createBooking(input: $input) {
         id
         status
         startsAt
         endsAt
      }
   }
   ```
   ```graphql
   mutation UpdateMyProfile($input: ProfileUpdateInput!) {
      updateProfile(input: $input) {
         id
         fullName
         phone
      }
   }
   ```

### Identity (`:8081`)
- `POST /auth/register` — создание пользователя
- `POST /auth/login` — возвращает токен и профиль
- `POST /auth/logout`, `POST /auth/refresh`
- `GET /users/me`, `PATCH /users/me` — работа с профилем через Bearer-токен
- `GET /internal/users/{id}`, `PATCH /internal/users/{id}`
- `POST /internal/auth/validate` — проверка токена для других сервисов

### Facility (`:8082`)
- `GET /facilities` — фильтры по городу/типу
- `POST /facilities`, `GET /facilities/{id}`, `PATCH`, `DELETE`
- `GET /facilities/{id}/availability`, `PUT /facilities/{id}/availability`

### Booking (`:8083`)
- `GET /bookings?facility_id=&user_id=`
- `POST /bookings` — валидация с Facility + проверка пересечений
- `GET /bookings/{id}`, `PATCH`, `DELETE`
- `POST /bookings/{id}/confirm`, `/cancel`
- `GET /availability/search?facility_id=&date=YYYY-MM-DD`

### Notification (`:8084`)
- `GET /notifications`
- `POST /notifications/send`
- `POST /notifications/test`
- `GET /notifications/{id}`

## PostgreSQL и миграции
- SQL-скрипты лежат в `services/<service>/migrations/001_init.sql` и могут применяться любым менеджером миграций.
- Identity автоматически докидывает демо-пользователей, если таблица пуста, Facility — пару площадок, Booking стартует с пустой таблицей.
- В production-режиме стоит переключить DSN на управляемую СУБД и добавить контроль прав.

## Что дальше
- Расширить GraphQL-схему (мутации, авторизация, батчинг запросов)
- Автоматизировать применение миграций/сидов внутри контейнеров (init job или entrypoint)
- Добавить авторизацию ролей на шлюзе и в доменных сервисах

Прототип уже можно отдавать фронтенду — все ручки работают и возвращают наглядный JSON.
