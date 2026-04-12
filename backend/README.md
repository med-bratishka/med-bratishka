# Medbratishka - Медицинская платформа связи врач-пациент

## Описание

Платформа для связи между докторами и пациентами, управления клиниками и их администраторами.

## Архитектура БД

### Основные таблицы:

1. **users** - таблица пользователей системы
   - Содержит данные всех участников (админы, врачи, пациенты)
   - Роль определяется полем `role`
   - Все пользователи требуют верификации

2. **auth_tokens** - таблица хранения токенов
   - Хранит secret каждого сеанса пользователя
   - Поддерживает множественные сеансы (session_number)
   - Токены могут быть отозваны

3. **clinics** - таблица клиник

4. **clinic_admins** - связь между клиниками и администраторами

5. **doctor_profiles** - профили врачей
   - Информация о специализации
   - Номер лицензии

6. **doctor_clinic_memberships** - членство врача в клинике
   - Статусы: pending, active, suspended, rejected
   - Врач приглашается админом

7. **doctor_patients** - связь между врачом и пациентом

8. **patient_profiles** - профили пациентов

9. **chats** - чаты между врачом и пациентом

10. **messages** - сообщения в чатах

## Установка

### Требования

- Go 1.25.2+
- PostgreSQL 12+
- Goose (для миграций)

### Установка зависимостей

```bash
go mod download
go mod tidy
```

### Установка Goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Создание базы данных

```bash
createdb medbratishka
```

### Запуск миграций

```bash
goose -dir ./migrations postgres "user=postgres password=postgres dbname=medbratishka sslmode=disable" up
```

## Конфигурация

Приложение поддерживает два способа конфигурации:

1. JSON-файл окружения (`configs/<env>.json`)
2. Переменные окружения (имеют приоритет над JSON)

Выбор файла:

- через `CONFIG_PATH` (абсолютный/относительный путь),
- либо через `APP_ENV` (по умолчанию `local`), тогда путь будет `configs/<APP_ENV>.json`.

Пример локального файла: `configs/local.json`.

Поддерживаемые ENV override:

```bash
# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=medbratishka
DB_SSLMODE=disable
DB_CERT_LOC=

# Auth
JWT_SECRET=your-secret-key-change-in-production
ACCESS_TTL=15m
REFRESH_TTL=168h

# S3
S3_ENDPOINT=localhost:9000
S3_REGION=us-east-1
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=medbratishka
S3_USE_SSL=false
S3_MAX_UPLOAD_SIZE_MB=15
```

## Запуск через Docker Compose

Готовый compose-файл находится в `deployments/docker-compose.yml`.

```bash
cd deployments
docker compose up --build -d
```

Проверка:

```bash
docker compose ps
docker compose logs -f app
```

Остановка:

```bash
docker compose down
```

Для очистки томов БД/S3:

```bash
docker compose down -v
```

## Локальный запуск без Docker

```bash
APP_ENV=local go run cmd/server/main.go
```

## API Endpoints

### Аутентификация

#### Регистрация
```
POST /auth/register
Content-Type: application/json

{
  "login": "john.doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "middle_name": "Michael",
  "role": "doctor"
}

Response:
{
  "access_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1712973600000
  },
  "refresh_token": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1720753600000
  },
  "server_time": 1712973600000,
  "user": {
    "id": 1,
    "role": "doctor",
    "login": "john.doe",
    "email": "john@example.com",
    "is_verified": false
  }
}
```

#### Вход
```
POST /auth/login
Content-Type: application/json

{
  "access_parameter": "john.doe",  // может быть login, email или phone
  "password": "securePassword123"
}

Response: (аналогично регистрации)
```

#### Обновление токенов
```
POST /auth/refresh
Authorization: Bearer {refresh_token}

Response: (новая пара access и refresh токенов)
```

#### Выход (текущий сеанс)
```
POST /auth/logout
Authorization: Bearer {access_token}

Response:
{
  "success": true
}
```

#### Выход из всех сеансов
```
POST /auth/full-logout
Authorization: Bearer {access_token}

Response:
{
  "success": true
}
```

## Структура проекта

```
medbratishka/
├── cmd/
│   └── server/
│       └── main.go           # Точка входа приложения
├── internal/
│   └── repository/
│       ├── users.go          # Репозиторий пользователей
│       └── tokens.go         # Репозиторий токенов
├── migrations/               # SQL миграции (Goose)
├── pkg/
│   ├── config/
│   │   └── config.go         # Конфигурация приложения
│   ├── database/
│   │   └── connection.go     # Подключение к БД
│   ├── logger/
│   │   └── logger.go         # Логирование
│   └── time_manager/
│       └── time_manager.go   # Управление временем
└── portable-auth/
    ├── domain/
    │   └── models.go         # Доменные модели
    ├── handler/
    │   ├── auth.go           # HTTP обработчики
    │   └── middleware.go     # Middleware для авторизации
    ├── pkg/
    │   └── token/
    │       └── token.go      # JWT токены
    ├── repository/
    │   └── interfaces.go     # Интерфейсы репозиториев
    └── service/
        ├── auth.go           # Сервис аутентификации
        └── registration.go   # Сервис регистрации
```

## Роли пользователей

- **admin** - администратор клиники
- **doctor** - врач
- **patient** - пациент

## Требования к безопасности

⚠️ **Важно для production:**

1. Используйте bcrypt для хеширования паролей (не простое сравнение)
2. Генерируйте надежный JWT_SECRET
3. Используйте SSL/TLS для всех соединений
4. Добавьте rate limiting на эндпоинты аутентификации
5. Реализуйте двухфакторную аутентификацию
6. Добавьте верификацию email/phone

## Следующие шаги

- [ ] Реализовать bcrypt для хеширования паролей
- [ ] Добавить верификацию email
- [ ] Реализовать управление профилями врачей
- [ ] Реализовать управление клиниками
- [ ] Добавить эндпоинты для работы с чатами
- [ ] Добавить WebSocket для real-time сообщений
- [ ] Реализовать логирование
- [ ] Добавить rate limiting
- [ ] Настроить CORS
- [ ] Добавить тесты

