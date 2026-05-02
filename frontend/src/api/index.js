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
}

export const chatApi = {
  getChats: () => api.get('/chats'),
  createChat: (doctorId) => api.post('/chats', { doctor_id: doctorId }),
  getMessages: (chatId, params = {}) =>
    api.get(`/chats/${chatId}/messages`, { params }),
  sendMessage: (chatId, content) =>
    api.post(`/chats/${chatId}/messages`, { content }),
}

export const doctorApi = {
  setCode: (code) => api.put('/doctors/me/code', { doctor_code: code }),
}

export const patientApi = {
  bindDoctor: (code) => api.post('/patients/me/bind-doctor', { doctor_code: code }),
}

export const adminApi = {
  bindDoctorToClinic: (clinicId, doctorId) =>
    api.post(`/clinics/${clinicId}/doctors/${doctorId}/bind`),
  // Получить всех пользователей через health + users из БД — используем /chats как proxy
  // Реального GET /users нет в API, поэтому получаем через chats
  getUsers: () => api.get('/chats'),
}

export const healthApi = {
  check: () => api.get('/health'),
}

export default api
