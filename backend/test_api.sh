#!/bin/bash

# Примеры CURL команд для тестирования API Medbratishka

BASE_URL="http://localhost:8080"

echo "========================================="
echo "API Testing Examples - Medbratishka"
echo "========================================="
echo ""

# 1. Health Check
echo "1. Health Check"
echo "----------------------------------------"
curl -X GET "$BASE_URL/health" 2>/dev/null | jq . || echo "Server не запущен"
echo ""
echo ""

# 2. Register Doctor
echo "2. Регистрация врача"
echo "----------------------------------------"
DOCTOR_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "doctor.ivanov",
    "email": "ivanov@clinic.com",
    "phone": "+79991234567",
    "password": "SecurePass123!",
    "first_name": "Иван",
    "last_name": "Иванов",
    "middle_name": "Иванович",
    "role": "doctor"
  }')

echo "$DOCTOR_RESPONSE" | jq .
DOCTOR_ACCESS_TOKEN=$(echo "$DOCTOR_RESPONSE" | jq -r '.access_token.token')
DOCTOR_ID=$(echo "$DOCTOR_RESPONSE" | jq -r '.user.id')
echo "Doctor ID: $DOCTOR_ID"
echo "Doctor Access Token: ${DOCTOR_ACCESS_TOKEN:0:30}..."
echo ""
echo ""

# 3. Register Patient
echo "3. Регистрация пациента"
echo "----------------------------------------"
PATIENT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "patient.petrov",
    "email": "petrov@patient.com",
    "phone": "+79997654321",
    "password": "PatientPass123!",
    "first_name": "Петр",
    "last_name": "Петров",
    "middle_name": "Петрович",
    "role": "patient"
  }')

echo "$PATIENT_RESPONSE" | jq .
PATIENT_ACCESS_TOKEN=$(echo "$PATIENT_RESPONSE" | jq -r '.access_token.token')
PATIENT_REFRESH_TOKEN=$(echo "$PATIENT_RESPONSE" | jq -r '.refresh_token.token')
PATIENT_ID=$(echo "$PATIENT_RESPONSE" | jq -r '.user.id')
echo "Patient ID: $PATIENT_ID"
echo "Patient Access Token: ${PATIENT_ACCESS_TOKEN:0:30}..."
echo ""
echo ""

# 4. Login Test
echo "4. Вход в систему"
echo "----------------------------------------"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "access_parameter": "doctor.ivanov",
    "password": "SecurePass123!"
  }')

echo "$LOGIN_RESPONSE" | jq .
echo ""
echo ""

# 5. Refresh Tokens
echo "5. Обновление токенов (используя refresh token)"
echo "----------------------------------------"
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Authorization: Bearer $PATIENT_REFRESH_TOKEN")

echo "$REFRESH_RESPONSE" | jq .
NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token.token')
echo "New Access Token: ${NEW_ACCESS_TOKEN:0:30}..."
echo ""
echo ""

# 6. Full Logout
echo "6. Выход из всех сеансов"
echo "----------------------------------------"
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/full-logout" \
  -H "Authorization: Bearer $DOCTOR_ACCESS_TOKEN")

echo "$LOGOUT_RESPONSE" | jq .
echo ""
echo ""

# 7. Error Handling - Invalid Token
echo "7. Ошибка - Неверный токен"
echo "----------------------------------------"
ERROR_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer invalid_token")

echo "$ERROR_RESPONSE" | jq . || echo "$ERROR_RESPONSE"
echo ""
echo ""

# 8. Error Handling - Missing Fields
echo "8. Ошибка - Отсутствующие поля"
echo "----------------------------------------"
ERROR_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "test.user",
    "password": "test123"
  }')

echo "$ERROR_RESPONSE"
echo ""
echo ""

echo "========================================="
echo "Testing Complete"
echo "========================================="

