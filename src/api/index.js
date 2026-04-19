import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:8080',
  headers: { 'Content-Type': 'application/json' },
})

// Подставляем access_token в каждый запрос
api.interceptors.request.use((config) => {
  const stored = localStorage.getItem('medcare_user')
  if (stored) {
    const { token } = JSON.parse(stored)
    if (token) config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Если токен протух (401) — разлогиниваем
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

  register: (data) =>
    api.post('/auth/register', data),

  refresh: (refreshToken) =>
    api.post('/auth/refresh', null, {
      headers: { Authorization: `Bearer ${refreshToken}` },
    }),

  logout: () => api.post('/auth/logout'),
}

export const patientsApi = {
  getAll: () => api.get('/doctor/patients'),
  getById: (id) => api.get(`/doctor/patients/${id}`),
}

export const chatApi = {
  getMessages: (patientId) => api.get(`/chat/${patientId}`),
  sendMessage: (patientId, text) => api.post(`/chat/${patientId}`, { text }),
}

export const medsApi = {
  getByPatient: (patientId) => api.get(`/medications/${patientId}`),
  add: (patientId, med) => api.post(`/medications/${patientId}`, med),
  update: (medId, data) => api.put(`/medications/${medId}`, data),
  delete: (medId) => api.delete(`/medications/${medId}`),
}

export const inviteApi = {
  generate: () => api.post('/invite/generate'),
  getAll: () => api.get('/invite/list'),
  use: (code) => api.post('/invite/use', { code }),
}

export const remindersApi = {
  getToday: () => api.get('/reminders/today'),
}

export default api
