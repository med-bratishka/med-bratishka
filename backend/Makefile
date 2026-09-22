.PHONY: help build run run-full generate generate-local build-models build-models-ci postgres.start postgres.stop migrate.up migrate.down migrate.create migrate.status migrate.reset swagger.start swagger.stop lint fmt test deps clean doctor version setup

PROJECT_DIR = $(shell pwd)
PROJECT_NAME = medbratishka
DB_HOST = localhost
DB_PORT = 5432
DB_USER = postgres
DB_PASSWORD = postgres
DB_NAME = medbratishka
DB_SSLMODE = disable
SWAGGER_IMAGE = medbratishka-swagger
SWAGGER_SPEC = swagger.yaml
SWAGGER_UI_SPEC = openapi.yaml
SWAGGER_PORT = 8081

help:
	@echo "Medbratishka - Build Management"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build              - Компилировать приложение"
	@echo "  make run                - Запустить приложение"
	@echo "  make run-full           - Полная сборка (swagger + миграции + run)"
	@echo ""
	@echo "Database:"
	@echo "  make postgres.start     - Запустить PostgreSQL контейнер"
	@echo "  make postgres.stop      - Остановить PostgreSQL контейнер"
	@echo "  make migrate.up         - Выполнить миграции вверх"
	@echo "  make migrate.down       - Откатить миграции"
	@echo "  make migrate.status     - Показать статус миграций"
	@echo "  make migrate.create     - Создать новую миграцию (name=migration_name)"
	@echo ""
	@echo "Swagger & API:"
	@echo "  make generate           - Генерировать swagger models"
	@echo "  make swagger.start      - Запустить Swagger UI"
	@echo "  make swagger.stop       - Остановить Swagger UI"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint               - Запустить linter"
	@echo "  make fmt                - Отформатировать код"
	@echo ""
	@echo "Clean:"
	@echo "  make clean              - Очистить скомпилированные файлы"

# ============ BUILD & RUN ============

build:
	@echo "Компилирую приложение..."
	go build -o bin/server ./cmd/server/
	@echo "Компиляция завершена!"

run:
	@echo "Запускаю приложение..."
	go run ./cmd/server/main.go

run-full:
	@echo "Полная сборка проекта..."
	@echo "1. Генерирую Swagger модели..."
	@$(MAKE) generate
	@echo "2. Запускаю PostgreSQL..."
	@$(MAKE) postgres.start
	@echo "3. Выполняю миграции..."
	@$(MAKE) migrate.up
	@echo "4. Запускаю приложение..."
	@$(MAKE) run

# ============ SWAGGER & MODELS ============

generate:
	@echo "Генерирую Swagger модели..."
	@if [ ! -f "$(SWAGGER_SPEC)" ]; then \
		echo "Не найден файл спецификации: $(SWAGGER_SPEC)"; \
		echo "Создай swagger 2.0 spec в $(SWAGGER_SPEC)"; \
		exit 1; \
	fi
	docker build -f Dockerfile.swagger -t $(SWAGGER_IMAGE) .
	docker run --rm \
		--user $(shell id -u):$(shell id -g) \
		-v $(shell pwd):/app \
		-w /app \
		$(SWAGGER_IMAGE) generate model -f $(SWAGGER_SPEC)
	@echo "Модели успешно сгенерированы!"
	@echo "Файлы сгенерированы в: ./models/"

build-models: generate

build-models-ci:
	@echo "CI: Генерирую Swagger модели..."
	@if [ ! -f "$(SWAGGER_SPEC)" ]; then \
		echo "Не найден файл спецификации: $(SWAGGER_SPEC)"; \
		exit 1; \
	fi
	docker build -f Dockerfile.swagger -t $(SWAGGER_IMAGE) .
	docker run --rm \
		--user $(shell id -u):$(shell id -g) \
		-v $(shell pwd):/app \
		-w /app \
		$(SWAGGER_IMAGE) generate model -f $(SWAGGER_SPEC)

generate-local:
	@echo "Генерирую Swagger модели локально..."
	@if ! command -v swagger >/dev/null 2>&1; then \
		echo "Устанавливаю swagger..."; \
		go install github.com/go-swagger/go-swagger/cmd/swagger@latest; \
	fi
	swagger generate model -f $(SWAGGER_SPEC)
	@echo "Модели успешно сгенерированы!"

swagger.start:
	@echo "Запускаю Swagger UI..."
	@if [ ! -f "$(SWAGGER_UI_SPEC)" ]; then \
		echo "Не найден файл спецификации для UI: $(SWAGGER_UI_SPEC)"; \
		exit 1; \
	fi
	@if [ ! "$(shell docker ps -q -f name=medbratishka-swagger)" ]; then \
		docker run --restart unless-stopped -d \
			--name medbratishka-swagger \
			-p $(SWAGGER_PORT):8080 \
			-e SWAGGER_JSON=/specs/$(SWAGGER_UI_SPEC) \
			-v $(shell pwd):/specs \
			swaggerapi/swagger-ui; \
		echo "Swagger UI запущен на http://localhost:$(SWAGGER_PORT)"; \
	else \
		echo "Swagger UI уже запущен"; \
	fi

swagger.stop:
	@echo "Останавливаю Swagger UI..."
	-docker stop medbratishka-swagger
	-docker rm medbratishka-swagger
	@echo "Swagger UI остановлен"

# ============ DATABASE ============

postgres.start:
	@echo "Запускаю PostgreSQL контейнер..."
	@if [ ! "$(shell docker ps -q -f name=medbratishka-postgres)" ]; then \
		if [ "$(shell docker ps -aq -f status=exited -f name=medbratishka-postgres)" ]; then \
			docker rm medbratishka-postgres; \
		fi; \
		docker run --restart unless-stopped -d \
			-p $(DB_PORT):5432 \
			-e POSTGRES_USER=$(DB_USER) \
			-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
			-e POSTGRES_DB=$(DB_NAME) \
			--name medbratishka-postgres \
			postgres:15-alpine; \
		sleep 3; \
		echo "PostgreSQL запущен на localhost:$(DB_PORT)"; \
		echo "Создаю БД $(DB_NAME)..."; \
		docker exec medbratishka-postgres psql -U $(DB_USER) -c "CREATE DATABASE IF NOT EXISTS $(DB_NAME);" 2>/dev/null || true; \
		echo "БД готова к использованию"; \
	else \
		echo "PostgreSQL уже запущен"; \
	fi

postgres.stop:
	@echo "Останавливаю PostgreSQL..."
	-docker stop medbratishka-postgres
	-docker rm medbratishka-postgres
	@echo "PostgreSQL остановлен"

postgres.reset:
	@echo "Сброс БД..."
	docker exec medbratishka-postgres psql -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	docker exec medbratishka-postgres psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"
	@echo "БД сброшена"

# ============ MIGRATIONS ============

migrate.up:
	@echo "Выполняю миграции вверх..."
	goose -dir migrations postgres "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)" up
	@echo "Миграции выполнены"

migrate.down:
	@echo "Откатываю последнюю миграцию..."
	goose -dir migrations postgres "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)" down
	@echo "Миграция отката выполнена"

migrate.reset:
	@echo "Сброс всех миграций..."
	goose -dir migrations postgres "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)" reset
	@echo "Миграции сброшены"

migrate.status:
	@echo "Статус миграций:"
	goose -dir migrations postgres "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)" status

migrate.create:
	@if [ -z "$(name)" ]; then \
		echo "Ошибка: укажи название миграции"; \
		echo "Использование: make migrate.create name=migration_name"; \
		exit 1; \
	fi
	@echo "Создаю миграцию '$(name)'..."
	goose -dir migrations create $(name) sql
	@echo "Миграция создана в migrations/"

# ============ CODE QUALITY ============

lint:
	@echo "Запускаю linter..."
	golangci-lint run ./... -c .golangci.yml
	@echo "Lint завершен"

fmt:
	@echo "Форматирую код..."
	gofmt -s -w ./internal ./cmd ./pkg
	@echo "Форматирование завершено"

test:
	@echo "Запускаю тесты..."
	go test -v -cover ./...
	@echo "Тесты завершены"

# ============ UTILITIES ============

deps:
	@echo "Скачиваю зависимости..."
	go mod download
	go mod tidy
	@echo "Зависимости обновлены"

clean:
	@echo "Очищаю..."
	rm -f bin/server
	docker rmi $(SWAGGER_IMAGE) 2>/dev/null || true
	go clean
	@echo "Очистка завершена"

doctor:
	@echo "Проверяю окружение..."
	@command -v go >/dev/null 2>&1 && echo "Go установлен ($(shell go version | awk '{print $$3}'))" || echo "Go не установлен"
	@command -v docker >/dev/null 2>&1 && echo "Docker установлен" || echo "Docker не установлен"
	@command -v goose >/dev/null 2>&1 && echo "Goose установлен" || echo "Goose не установлен"
	@command -v golangci-lint >/dev/null 2>&1 && echo "golangci-lint установлен" || echo "golangci-lint не установлен"
	@echo ""
	@echo "Docker контейнеры:"
	@docker ps --filter "name=medbratishka" --format "table {{.Names}}\t{{.Status}}"
	@echo ""
	@echo "Проверка БД:"
	@if [ "$(shell docker ps -q -f name=medbratishka-postgres)" ]; then \
		echo "PostgreSQL запущен"; \
		docker exec medbratishka-postgres psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT 1;" 2>/dev/null && echo "БД доступна" || echo "БД недоступна"; \
	else \
		echo "PostgreSQL не запущен"; \
	fi

version:
	@echo "Medbratishka версия информация:"
	@go version
	@echo "Go modules версия:"
	@go mod graph | head -1

# ============ QUICK SETUP ============

setup:
	@echo "Инициализирую проект..."
	@$(MAKE) deps
	@$(MAKE) postgres.start
	@$(MAKE) migrate.up
	@echo "Проект готов к разработке!"
	@echo ""
	@echo "Для запуска используй:"
	@echo "  make run        - Запустить приложение"
	@echo "  make swagger.start - Открыть Swagger UI"
	@echo "  make help       - Показать все команды"

.DEFAULT_GOAL := help

