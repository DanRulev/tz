# 📌 Тестовое задание: REST-сервис агрегации подписок

Реализация REST-сервиса для агрегации данных об онлайн-подписках пользователей.

## 🧩 Функциональность

Сервис предоставляет следующие возможности:

- **CRUDL** операции над записями о подписках:
  - Создание (`POST /subscriptions`)
  - Чтение по ID (`GET /subscriptions/{id}`)
  - Обновление (`PATCH /subscriptions/{id}`)
  - Удаление (`DELETE /subscriptions/{id}`)
  - Список с фильтрацией и пагинацией (`GET /subscriptions`)
- Подсчёт суммарной стоимости подписок за указанный период с фильтрацией по пользователю и названию сервиса (`GET /subscriptions/cost`)
- Поддержка Swagger-документации (`GET /swagger/*`)
- Health-check эндпоинт (`GET /health`)

---

## 🛠️ Технологии

- **Язык**: Go 1.25+
- **Фреймворк**: [Gin](https://gin-gonic.com/)
- **База данных**: PostgreSQL
- **Миграции**: [migrate](https://github.com/golang-migrate/migrate)
- **Логирование**: [Zap](https://github.com/uber-go/zap)
- **Валидация**: [validator/v10](https://github.com/go-playground/validator)
- **Конфигурация**: [Viper](https://github.com/spf13/viper)
- **Контейнеризация**: Docker + Docker Compose

---

## 🚀 Запуск

### Требования

- Docker и Docker Compose
- Go 1.25+ (для локальной разработки)

### Быстрый старт с Docker

1. Убедитесь, что у вас есть файл .env в корне проекта. Пример содержимого:
```
DB_NAME=subscription_db
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_PORT=5432
DB_SSL=disable
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
```
2. Выполните:

```bash
make docker-up
```

Сервис будет доступен на http://localhost:8080.


### 📄 API Документация
Swagger UI доступен по адресу:
👉 http://localhost:8080/swagger/index.html

Документация сгенерирована с помощью swaggo .