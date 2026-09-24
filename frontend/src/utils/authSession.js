export function persistAuthResponse(data, login, fallbackName = '') {
  const { access_token, refresh_token, user, trusted_device_token } = data
  if (!access_token?.token || !refresh_token?.token || !user?.id) {
    throw new Error('auth response is incomplete')
  }
  if (trusted_device_token) {
    localStorage.setItem('medcare_trusted_device', trusted_device_token)
  }
  login({
    id: user.id,
    name: fallbackName || user.login || user.email,
    login: user.login,
    email: user.email,
    first_name: user.first_name,
    last_name: user.last_name,
    role: user.role,
    token: access_token.token,
    refreshToken: refresh_token.token,
  })
}
