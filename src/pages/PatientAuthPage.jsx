import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { authApi } from '../api/index'

export default function PatientAuthPage() {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    phone: '',
    inviteCode: '',
  })
  const [loading, setLoading] = useState(false)

  const { login, authError, setAuthError } = useAuth()
  const navigate = useNavigate()

  const handleChange = (e) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }))
    setAuthError(null)
  }

  const handleSubmit = async () => {
    setLoading(true)
    setAuthError(null)

    try {
      let res

      if (mode === 'login') {
        res = await authApi.login(form.email, form.password)
      } else {
        if (!form.firstName || !form.lastName || !form.phone || !form.inviteCode) {
          setAuthError('Заполните все поля')
          setLoading(false)
          return
        }
        res = await authApi.register({
          login: form.email,
          email: form.email,
          phone: form.phone,
          password: form.password,
          first_name: form.firstName,
          last_name: form.lastName,
          middle_name: '',
          role: 'patient',
          invite_code: form.inviteCode,
        })
      }

      const { access_token, refresh_token, user } = res.data

      login({
        id: user.id,
        name: `${form.firstName || ''} ${form.lastName || ''}`.trim() || user.login,
        role: user.role,
        token: access_token.token,
        refreshToken: refresh_token.token,
      })

      navigate('/patient')
    } catch (err) {
      const message =
        err?.response?.data?.message ||
        err?.response?.data?.error ||
        'Неверный email или пароль'
      setAuthError(message)
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') handleSubmit()
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
            <h1 className="text-base font-semibold text-gray-800">
              {mode === 'login' ? 'Добро пожаловать' : 'Регистрация пациента'}
            </h1>
            <p className="text-xs text-gray-400 mt-0.5">Личный кабинет пациента</p>
          </div>

          <div className="p-7 flex flex-col gap-4">

            {mode === 'register' && (
              <>
                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-gray-500">Имя</label>
                    <input
                      className="input-field"
                      name="firstName"
                      placeholder="Алёна"
                      value={form.firstName}
                      onChange={handleChange}
                      onKeyDown={handleKeyDown}
                    />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs text-gray-500">Фамилия</label>
                    <input
                      className="input-field"
                      name="lastName"
                      placeholder="Иванова"
                      value={form.lastName}
                      onChange={handleChange}
                      onKeyDown={handleKeyDown}
                    />
                  </div>
                </div>

                <div className="flex flex-col gap-1.5">
                  <label className="text-xs text-gray-500">Телефон</label>
                  <input
                    className="input-field"
                    name="phone"
                    type="tel"
                    placeholder="+79001234567"
                    value={form.phone}
                    onChange={handleChange}
                    onKeyDown={handleKeyDown}
                  />
                </div>
              </>
            )}

            <div className="flex flex-col gap-1.5">
              <label className="text-xs text-gray-500">Email</label>
              <input
                className="input-field"
                name="email"
                type="email"
                placeholder="patient@mail.ru"
                value={form.email}
                onChange={handleChange}
                onKeyDown={handleKeyDown}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs text-gray-500">Пароль</label>
              <input
                className="input-field"
                name="password"
                type="password"
                placeholder="••••••••"
                value={form.password}
                onChange={handleChange}
                onKeyDown={handleKeyDown}
              />
            </div>

            {mode === 'register' && (
              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-gray-500">Код приглашения от врача</label>
                <input
                  className="input-field font-mono tracking-widest uppercase"
                  name="inviteCode"
                  placeholder="MED-XXXXX"
                  value={form.inviteCode}
                  onChange={handleChange}
                  onKeyDown={handleKeyDown}
                />
              </div>
            )}

            {authError && (
              <p className="text-xs text-red-500 bg-red-50 rounded-lg px-3 py-2">{authError}</p>
            )}

            <button
              onClick={handleSubmit}
              disabled={loading}
              className="btn-primary w-full py-2.5 mt-1 disabled:opacity-50"
            >
              {loading ? 'Загрузка...' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
            </button>

            <div className="flex items-center gap-3">
              <div className="flex-1 h-px bg-gray-100" />
              <span className="text-xs text-gray-400">
                {mode === 'login' ? 'нет аккаунта?' : 'есть аккаунт?'}
              </span>
              <div className="flex-1 h-px bg-gray-100" />
            </div>

            <button
              onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setAuthError(null) }}
              className="btn-secondary w-full py-2.5 text-xs"
            >
              {mode === 'login' ? 'Зарегистрироваться' : 'Войти'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
