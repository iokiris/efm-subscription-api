# 🚀 EM Subscriptions API



> 🎯 **API для управления подписками и получения агрегированных сумм**

## ✨ Особенности

- 🏗️ **Clean Architecture** с четким разделением слоев
- 🐳 **Docker** контейнеризация с multi-stage сборкой
- 📊 **Мониторинг** с Prometheus, Grafana и Jaeger
- 🗄️ **PostgreSQL** с миграциями
- ⚡ **Redis** для кэширования
- 🐰 **RabbitMQ** для асинхронной обработки
- 📚 **Swagger** документация
- 🧪 **Тесты** с покрытием кода
- 🔒 **Безопасность** с Trivy сканированием
- 🚀 **CI/CD** с GitHub Actions

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3+

### Запуск с Docker Compose

```bash
# Клонируем репозиторий
git clone https://github.com/iokiris/efm-subscription-api.git
cd efm-subscription-api

# Создаем .env файл
cp .env.example .env

# Запускаем все сервисы
docker-compose -f docker-compose.dev.yaml up -d

# Проверяем статус
docker-compose -f docker-compose.dev.yaml ps
```

## 📚 API Документация

После запуска приложения Swagger документация доступна по адресу (порты по умолчанию):
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI Spec**: http://localhost:8080/swagger/doc.json

## 🧪 Тестирование

```bash
# Запуск всех тестов
go test -v ./...

# Запуск тестов с покрытием
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Просмотр покрытия
go tool cover -html=coverage.out -o coverage.html
```



## 🏗️ Архитектура

```
         ┌─────────────────┐
         │   HTTP Layer    │
         │  - Handlers     │
         │  - Middleware   │
         │  - Routes       │
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │  Business Layer │
         │  - Services     │
         │  - Interfaces   │
         │  - Logic        │
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │   Data Layer    │
         │  - Repositories │
         │  - Models       │
         │  - Database     │
         └────────┬────────┘
                  │
      ┌───────────┴───────────┐
      ▼                       ▼
┌───────────────┐       ┌───────────────┐
│ Infrastructure│       │   External    │
│ - Redis       │       │ - PostgreSQL  │
│ - RabbitMQ    │       │ - Redis       │
│ - Docker      │       │ - RabbitMQ    │
└───────────────┘       └───────────────┘
           │
           ▼
   ┌───────────────┐
   │  Monitoring   │
   │ - Prometheus  │
   │ - Grafana     │
   │ - Jaeger      │
   └───────────────┘

```

## 🚀 CI/CD

Проект использует GitHub Actions для автоматизации:

- **🧪 Тестирование** - автоматический запуск тестов
- **🔍 Линтинг** - проверка качества кода
- **🐳 Docker** - сборка и публикация образов
- **🔒 Безопасность** - сканирование уязвимостей
- **🚀 Деплой** - автоматический деплой в staging/production

