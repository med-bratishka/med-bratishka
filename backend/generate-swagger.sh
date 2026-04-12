#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "🔄 Генерирую swagger-модели через Makefile..."
make build-models

echo "✅ Готово"
echo "📁 Проверь директорию ./models/"
