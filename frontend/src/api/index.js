import axios from 'axios'

const api = axios.create({
  baseURL: '',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const stored = localStorage.getItem('medcare_user')
  if (stored) {
    const { token } = JSON.parse(stored)
    if (token) config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    const url = err.config?.url || ''
    const isAuthFlow = url.startsWith('/auth/login') || url.startsWith('/auth/2fa/verify')
    if (err.response?.status === 401 && !isAuthFlow) {
      localStorage.removeItem('medcare_user')
      window.location.href = '/auth'
    }
    return Promise.reject(err)
  }
)

export const authApi = {
  login: (email, password) => {
    const trustedDeviceToken = localStorage.getItem('medcare_trusted_device')
    return api.post('/auth/login', {
      access_parameter: email,
      password,
      trusted_device_token: trustedDeviceToken || undefined,
      device_name: navigator.userAgent,
    })
  },
  register: (data) => api.post('/auth/register', data),
  setup2FA: () => api.post('/auth/2fa/setup'),
  confirm2FA: (code) => api.post('/auth/2fa/confirm', { code }),
  verify2FA: ({ challengeId, code, recoveryCode, rememberDevice }) =>
    api.post('/auth/2fa/verify', {
      challenge_id: challengeId,
      code: recoveryCode ? undefined : code,
      recovery_code: recoveryCode || undefined,
      trust_device: !!rememberDevice,
      device_name: navigator.userAgent,
    }),
  disable2FA: (code) => api.post('/auth/2fa/disable', { code }),
  regenerateRecoveryCodes: (code) => api.post('/auth/2fa/recovery-codes', { code }),
  refresh: (refreshToken) =>
    api.post('/auth/refresh', null, {
      headers: { Authorization: `Bearer ${refreshToken}` },
    }),
  logout: () => api.post('/auth/logout'),
  logoutAll: () => api.post('/auth/logout-all'),
}

export const chatApi = {
  getChats: () => api.get('/chats'),
  createChat: (doctorId) => api.post('/chats', { doctor_id: doctorId }),
  closeChat: (chatId) => api.delete(`/chats/${chatId}`),
  getMessages: (chatId, params = {}) =>
    api.get(`/chats/${chatId}/messages`, { params }),
  sendMessage: (chatId, content) =>
    api.post(`/chats/${chatId}/messages`, { content }),
  markRead: (chatId, lastReadMessageId) =>
    api.post(`/chats/${chatId}/read`, lastReadMessageId ? { last_read_message_id: lastReadMessageId } : {}),
}

export const doctorApi = {
  setCode: (code) => api.put('/doctors/me/code', { doctor_code: code }),
  unlinkPatient: (patientId) => api.delete(`/doctors/me/patients/${patientId}`),
}

export const patientApi = {
  bindDoctor: (code) => api.post('/patients/me/bind-doctor', { doctor_code: code }),
}

export const prescriptionApi = {
  getForPatient: (patientId) => api.get('/prescriptions', { params: { patient_id: patientId } }),
  create: (data) => api.post('/prescriptions', data),
  delete: (id) => api.delete(`/prescriptions/${id}`),
}

export const adminApi = {
  bindDoctorToClinic: (clinicId, doctorId) =>
    api.post(`/clinics/${clinicId}/doctors/${doctorId}/bind`),
  getUsers: () => api.get('/chats'),
}

export const healthApi = {
  check: () => api.get('/health'),
}

export default api
