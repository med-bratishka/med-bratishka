import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// Автоматически добавляем токен из localStorage
api.interceptors.request.use((config) => {
  const stored = localStorage.getItem('medcare_user')
  if (stored) {
    const { token } = JSON.parse(stored)
    if (token) config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ──── Auth ────
export const authApi = {
  login: (email, password, role) =>
    api.post('/auth/login', { email, password, role }),
  register: (data) =>
    api.post('/auth/register', data),
}

// ──── Пациенты (для врача) ────
export const patientsApi = {
  getAll: () => api.get('/doctor/patients'),
  getById: (id) => api.get(`/doctor/patients/${id}`),
}

// ──── Чат ────
export const chatApi = {
  getMessages: (patientId) => api.get(`/chat/${patientId}`),
  sendMessage: (patientId, text) =>
    api.post(`/chat/${patientId}`, { text }),
}

// ──── Лекарства ────
export const medsApi = {
  getByPatient: (patientId) => api.get(`/medications/${patientId}`),
  add: (patientId, med) => api.post(`/medications/${patientId}`, med),
  update: (medId, data) => api.put(`/medications/${medId}`, data),
  delete: (medId) => api.delete(`/medications/${medId}`),
}

// ──── Инвайт-коды ────
export const inviteApi = {
  generate: () => api.post('/invite/generate'),
  getAll: () => api.get('/invite/list'),
  use: (code) => api.post('/invite/use', { code }),
}

// ──── Напоминания ────
export const remindersApi = {
  getToday: () => api.get('/reminders/today'),
}

export default api
