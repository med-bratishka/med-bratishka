import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
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


export const authApi = {
  login: (email, password, role) =>
    api.post('/auth/login', { email, password, role }),
  register: (data) =>
    api.post('/auth/register', data),
}


export const patientsApi = {
  getAll: () => api.get('/doctor/patients'),
  getById: (id) => api.get(`/doctor/patients/${id}`),
}


export const chatApi = {
  getMessages: (patientId) => api.get(`/chat/${patientId}`),
  sendMessage: (patientId, text) =>
    api.post(`/chat/${patientId}`, { text }),
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
