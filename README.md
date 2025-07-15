# Документация API для сервиса аутентификации

## Конфигурация и запуск

### Переменные окружения

Создайте файл `.env` в корневой директории со следующим содержимым:

```
PORT=8080
DB_URL=postgres://postgres:postgres@postgres:5432/auth?sslmode=disable
TOKEN_SECRET=ваш_секретный_ключ_здесь
WEBHOOK_URL=http://ваш-вебхук-сервис/notify
```

**Пояснения:**
- `PORT`: Порт, на котором будет работать сервис аутентификации
- `DB_URL`: Строка подключения к базе данных PostgreSQL
- `TOKEN_SECRET`: Секретный ключ для генерации JWT токенов (используйте сильную случайную строку)
- `WEBHOOK_URL`: URL для отправки уведомлений о безопасности

### Запуск с помощью Docker

1. Убедитесь, что у вас установлены Docker и Docker Compose
2. Создайте файл `.env` как описано выше
3. Выполните команду в корне проекта:

```bash
docker compose -f docker-compose.yml up -d
```

Эта команда:
- Соберет образ сервиса аутентификации
- Запустит базу данных PostgreSQL
- Инициализирует базу данных с необходимыми таблицами
- Свяжет сервисы через Docker сеть
- Запустит все в фоновом режиме (`-d`)

Для остановки сервисов:

```bash
docker compose -f docker-compose.yml down
```

---

## Документация API

### 1. Регистрация пользователя

**Endpoint**: `POST /register`

**Описание**: Создает новую учетную запись пользователя

**Тело запроса**:
```json
{
  "Email": "user@example.com",
  "Password": "secure_password123"
}
```

**Успешный ответ** (200 OK):
```json
{
  "user": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "email": "user@example.com",
    "created_at": "2025-07-15T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4gZXhhbXBsZQ=="
}
```

**Возможные ошибки**:
- 400 Bad Request: Неверный формат запроса или ошибка хеширования пароля
- 500 Internal Server Error: Ошибка базы данных или ошибка генерации токена

### 2. Получение ID пользователя

**Endpoint**: `GET /users/id`

**Описание**: Возвращает ID текущего пользователя из его JWT токена

**Заголовки**:
```
Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...
```

**Успешный ответ** (200 OK):
```json
{
  "user_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6"
}
```

**Возможные ошибки**:
- 401 Unauthorized: Отсутствует или недействительный заголовок авторизации, истекший токен

### 3. Вход в систему

**Endpoint**: `POST /login`

**Описание**: Аутентифицирует пользователя и предоставляет токены доступа

**Тело запроса**:
```json
{
  "Email": "user@example.com",
  "Password": "secure_password123"
}
```

**Успешный ответ** (200 OK):
```json
{
  "user": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "email": "user@example.com",
    "created_at": "2025-07-15T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4gZXhhbXBsZQ=="
}
```

**Возможные ошибки**:
- 400 Bad Request: Неверный формат запроса или неправильный пароль
- 500 Internal Server Error: Ошибка базы данных или ошибка генерации токена

### 4. Обновление токенов

**Endpoint**: `POST /refresh`

**Описание**: Обменивает refresh token на новую пару access и refresh токенов

**Тело запроса**:
```json
{
  "RefreshToken": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4gZXhhbXBsZQ==",
  "AccessToken": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9..."
}
```

**Успешный ответ** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "ZGlmZmVyZW50IHJlZnJlc2ggdG9rZW4gZXhhbXBsZQ=="
}
```

**Возможные ошибки**:
- 400 Bad Request: Неверный формат запроса
- 401 Unauthorized: Недействительный или истекший refresh token, несовпадение пары токенов, изменение User-Agent
- 500 Internal Server Error: Ошибка базы данных или ошибка генерации токенов

**Примечания по безопасности**:
- При изменении User-Agent все refresh токены пользователя будут аннулированы
- При изменении IP-адреса будет отправлено уведомление на webhook URL

### 5. Выход из системы

**Endpoint**: `POST /logout`

**Описание**: Аннулирует все refresh токены для пользователя

**Заголовки**:
```
Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...
```

**Успешный ответ** (200 OK):
```json
{
  "message": "successfully logged out"
}
```

**Возможные ошибки**:
- 401 Unauthorized: Отсутствует или недействительный заголовок авторизации
- 400 Bad Request: Недействительный ID пользователя в токене
- 500 Internal Server Error: Ошибка базы данных

---
