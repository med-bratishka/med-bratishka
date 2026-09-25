import axios from 'axios'

const api = axios.create({
  baseURL: '',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const stored = localStorage.getItem('medcare_user')
  if (stored) {
    try {
      const { token } = JSON.parse(stored)
      if (token) config.headers.Authorization = `Bearer ${token}`
    } catch {}
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('medcare_user')
      window.location.href = '/auth'
    }
    return Promise.reject(err)
  }
)

export const authApi = {
  login: (email, password) =>
    api.post('/auth/login', { access_parameter: email, password }),
  register: (data) => api.post('/auth/register', data),
  refresh: (refreshToken) =>
    api.post('/auth/refresh', null, {
      headers: { Authorization: `Bearer ${refreshToken}` },
    }),
  logout: () => api.post('/auth/logout'),
  logoutAll: () => api.post('/auth/logout-all'),
  setup2FA: () => api.post('/auth/2fa/setup'),
  confirm2FA: (code) => api.post('/auth/2fa/confirm', { code }),
  disable2FA: (code) => api.post('/auth/2fa/disable', { code }),
  regenerateRecoveryCodes: (code) => api.post('/auth/2fa/recovery-codes', { code }),
  verifyTwoFactor: (challengeId, code) =>
    api.post('/auth/2fa/verify', { challenge_id: challengeId, code }),
}

export const chatApi = {
  getChats: () => api.get('/chats'),
  createChat: (doctorId) => api.post('/chats', { doctor_id: doctorId }),
  closeChat: (chatId) => api.delete(`/chats/${chatId}`),
  getMessages: (chatId, params = {}) =>
    api.get(`/chats/${chatId}/messages`, { params }),
  sendMessage: (chatId, content) =>
    api.post(`/chats/${chatId}/messages`, { content }),
  markRead: (chatId) => api.post(`/chats/${chatId}/read`),
}

export const doctorApi = {
  setCode: (code) => api.put('/doctors/me/code', { doctor_code: code }),
  unlinkPatient: (patientId) => api.delete(`/doctors/me/patients/${patientId}`),
}

export const patientApi = {
  bindDoctor: (code) => api.post('/patients/me/bind-doctor', { doctor_code: code }),
}

export const catalogApi = {
  // GET /doctors — список всех врачей (публичный)
  getDoctors: () => api.get('/doctors'),
  // GET /clinics — список всех клиник
  getClinics: () => api.get('/clinics'),
  // GET /clinics/:id — клиника по id
  getClinic: (id) => api.get(`/clinics/${id}`),
  // GET /doctors/:id — врач по id
  getDoctor: (id) => api.get(`/doctors/${id}`),
}

export const adminApi = {
  // Привязка врача к клинике
  bindDoctorToClinic: (clinicId, doctorId) =>
    api.post(`/clinics/${clinicId}/doctors/${doctorId}/bind`),
}

export const healthApi = {
  check: () => api.get('/health'),
}

export default api
