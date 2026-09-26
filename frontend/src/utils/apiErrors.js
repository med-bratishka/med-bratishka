export function getApiError(err) {
  const status = err?.response?.status ?? 0
  const data = err?.response?.data ?? {}
  return {
    status,
    code: data.code || data.error || '',
    message: data.message || data.error || '',
    details: data.details,
    isNetworkError: !err?.response,
  }
}

export function getLoginErrorMessage(err) {
  const { status, code, message, isNetworkError } = getApiError(err)

  if (isNetworkError) {
    return 'Backend недоступен. Проверьте, что сервер запущен и frontend проксирует запросы на API.'
  }
  if (status === 401 || code === 'INVALID_CREDENTIALS') {
    return 'Неверный email или пароль'
  }
  if (status === 400 || code === 'VALIDATION_ERROR') {
    return 'Проверьте email и пароль'
  }
  if (status === 404) {
    return 'Auth API недоступен. Проверьте proxy /auth и что backend запущен на 8080.'
  }
  if (status >= 500) {
    return 'Ошибка сервера авторизации. Попробуйте ещё раз позже.'
  }

  return message || 'Не удалось войти. Попробуйте ещё раз.'
}

export function getTwoFactorErrorMessage(err) {
  const { status, code, message, isNetworkError } = getApiError(err)

  if (isNetworkError) {
    return 'Backend недоступен. Проверьте подключение и попробуйте ещё раз.'
  }
  if (code === 'INVALID_2FA_CODE') {
    return 'Неверный код'
  }
  if (code === 'CHALLENGE_EXPIRED') {
    return 'Срок проверки истек, войдите заново'
  }
  if (code === 'CHALLENGE_NOT_FOUND') {
    return 'Проверка не найдена или уже использована, войдите заново'
  }
  if (code === 'TOO_MANY_ATTEMPTS' || status === 429) {
    return 'Слишком много попыток, войдите заново'
  }
  if (status === 400) {
    return 'Проверьте формат кода'
  }
  if (status === 404) {
    return '2FA API недоступен. Проверьте proxy /auth и что backend запущен на 8080.'
  }
  if (status >= 500) {
    return 'Ошибка сервера при проверке кода. Попробуйте ещё раз позже.'
  }

  return message || 'Не удалось проверить код'
}

export function getSecurityActionErrorMessage(err) {
  const { status, code, message, isNetworkError } = getApiError(err)

  if (isNetworkError) {
    return 'Backend недоступен. Проверьте подключение и попробуйте ещё раз.'
  }
  if (code === 'INVALID_2FA_CODE') {
    return 'Неверный код'
  }
  if (code === 'TOTP_ALREADY_ENABLED') {
    return 'Двухфакторная аутентификация уже включена'
  }
  if (code === 'TOTP_NOT_SETUP') {
    return 'Сначала начните настройку 2FA'
  }
  if (status === 401) {
    return 'Сессия истекла. Войдите заново.'
  }
  if (status === 404) {
    return '2FA API недоступен. Проверьте proxy /auth и что backend запущен на 8080.'
  }
  if (status >= 500) {
    return 'Ошибка сервера. Попробуйте ещё раз позже.'
  }

  return message || 'Не удалось выполнить действие'
}
