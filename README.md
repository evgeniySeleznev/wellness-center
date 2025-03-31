# Wellness Center Management System

Backend система для учета клиентов оздоровительного центра с интеграцией Kafka, Redis и Elasticsearch.

## 🔍 Основной функционал
- Регистрация клиентов с детальной анкетой
- Поиск по клиентской базе (Elasticsearch)
- Асинхронная обработка событий (Kafka)
- Кеширование данных (Redis)
- Healthcheck системы

## 🛠 Технологический стек
- **Язык**: Go 1.23
- **База данных**: PostgreSQL
- **Брокер сообщений**: Kafka
- **Поиск**: Elasticsearch
- **Кеш**: Redis
- **Контейнеризация**: Docker

## 🚀 Быстрый старт
```bash
git clone https://github.com/evgeniySeleznev/wellness-center.git
cd wellness-center
docker-compose up -d
```

Документация API доступна после запуска:
```
http://localhost:8080/api/v1/health
```

## 📝 Особенности реализации
- Graceful shutdown
- Модульная архитектура
- Интеграционные тесты
- Конфигурация через environment variables

## 📄 Лицензия
MIT
