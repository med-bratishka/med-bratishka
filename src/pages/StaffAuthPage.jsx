import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const STAFF_ROLES = [
  { value: 'doctor', label: 'Врач' },
  { value: 'admin', label: 'Администратор' },
]

export default function StaffAuthPage() {
  const [role, setRole] = useState('doctor')
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({ email: '', password: '', name: '' })

  const { login } = useAuth()
  const navigate = useNavigate()

  const handleChange = (e) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
  }

  // TODO: убрать мок после подключения бэка
  const handleSubmit = () => {
    const mockUser = {
      id: 1,
      name: form.name || (role === 'admin' ? 'Администратор' : 'Д-р Иванов'),
      role,
      token: 'mock-token',
    }
    login(mockUser)
    navigate(role === 'admin' ? '/admin' : '/doctor')
  }

  return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-slate-100 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <div className="flex items-center gap-2 mb-6 justify-center">
            <div className="w-8 h-8 bg-brand-400 rounded-lg flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path d="M8 2v12M2 8h12" stroke="white" strokeWidth="2.5" strokeLinecap="round"/>
              </svg>
            </div>
            <span className="text-xl font-medium text-gray-800">МедБратишка</span>
          </div>

          <div className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden">
            <div className="px-7 pt-6 pb-4 border-b border-gray-100">
              <h1 className="text-base font-semibold text-gray-800">Вход для персонала</h1>
              <p className="text-xs text-gray-400 mt-0.5">Портал врачей и администраторов</p>
            </div>

            <div className="p-7 flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-gray-500">Роль</label>
                <div className="grid grid-cols-2 gap-2">
                  {STAFF_ROLES.map((r) => (
                      <button
                          key={r.value}
                          type="button"
                          onClick={() => setRole(r.value)}
                          className={`py-2 rounded-lg text-sm font-medium border transition-colors ${
                              role === r.value
                                  ? 'bg-brand-50 text-brand-600 border-brand-300'
                                  : 'bg-white text-gray-400 border-gray-200 hover:bg-gray-50'
                          }`}
                      >
                        {r.label}
                      </button>
                  ))}
                </div>
              </div>

              {mode === 'register' && (
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-gray-500">Имя</label>
                    <input className="input-field" name="name" placeholder="Иван Иванов" value={form.name} onChange={handleChange} />
                  </div>
              )}

              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-gray-500">Email</label>
                <input className="input-field" name="email" type="email" placeholder="doctor@clinic.ru" value={form.email} onChange={handleChange} />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-gray-500">Пароль</label>
                <input className="input-field" name="password" type="password" placeholder="••••••••" value={form.password} onChange={handleChange} />
              </div>

              <button onClick={handleSubmit} className="btn-primary w-full py-2.5 mt-1">
                {mode === 'login' ? `Войти как ${role === 'admin' ? 'администратор' : 'врач'}` : 'Зарегистрироваться'}
              </button>

              <div className="flex items-center gap-3">
                <div className="flex-1 h-px bg-gray-100" />
                <span className="text-xs text-gray-400">{mode === 'login' ? 'нет аккаунта?' : 'есть аккаунт?'}</span>
                <div className="flex-1 h-px bg-gray-100" />
              </div>

              <button onClick={() => setMode(mode === 'login' ? 'register' : 'login')} className="btn-secondary w-full py-2.5 text-xs">
                {mode === 'login' ? 'Зарегистрироваться' : 'Войти'}
              </button>
            </div>
          </div>

          <p className="text-center text-xs text-gray-400 mt-4">
            Вы пациент?{' '}
            <a href={import.meta.env.VITE_PATIENT_APP_URL || '#'} className="text-brand-500 hover:underline">
              Перейти на портал пациентов
            </a>
          </p>
        </div>
      </div>
  )
}