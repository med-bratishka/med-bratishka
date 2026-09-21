#!/bin/bash

# Medbratishka Setup и Running Script

set -e

echo "================================"
echo "Medbratishka Project Setup"
echo "================================"
echo ""

# Check for PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL не установлен. Пожалуйста, установите PostgreSQL 12+"
    exit 1
fi

echo "✓ PostgreSQL найден"

# Check for Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Пожалуйста, установите Go 1.25.2+"
    exit 1
fi

echo "✓ Go найден"

# Create database
echo ""
echo "Создание базы данных..."
createdb medbratishka 2>/dev/null || echo "База данных уже существует"

# Install goose
echo ""
echo "Установка Goose..."
go install github.com/pressly/goose/v3/cmd/goose@latest

# Get dependencies
echo ""
echo "Загрузка зависимостей..."
go mod download
go mod tidy

# Run migrations
echo ""
echo "Запуск миграций..."
goose -dir ./migrations postgres "user=postgres password=postgres dbname=medbratishka sslmode=disable" up

echo ""
echo "================================"
echo "✓ Setup завершен успешно!"
echo "================================"
echo ""
echo "Для запуска сервера выполните:"
echo "  go run cmd/server/main.go"
echo ""
echo "Или с переменными окружения:"
echo "  JWT_SECRET=your-secret go run cmd/server/main.go"
echo ""
echo "Сервер будет доступен по адресу:"
echo "  http://localhost:8080"
echo ""
echo "Health check:"
echo "  curl http://localhost:8080/health"
echo ""
echo "Тестирование регистрации:"
echo "  curl -X POST http://localhost:8080/auth/register \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"login\":\"doctor.test\",\"password\":\"test123\",\"first_name\":\"Test\",\"last_name\":\"Doctor\",\"role\":\"doctor\"}'"
echo ""

