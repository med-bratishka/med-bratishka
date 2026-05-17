import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { authApi } from '../api/index'
import TwoFactorChallenge from '../components/auth/TwoFactorChallenge'
import { persistAuthResponse } from '../utils/authSession'

export default function PatientAuthPage() {
  const [mode, setMode] = useState('login')
  const [form, setForm] = useState({
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    phone: '',
  })
  const [loading, setLoading] = useState(false)
  const [fieldErrors, setFieldErrors] = useState({})
  const [pendingTwoFactor, setPendingTwoFactor] = useState(null)

  const { login, authError, setAuthError } = useAuth()
  const navigate = useNavigate()

  const handleChange = (e) => {
    const { name, value } = e.target
    setForm((prev) => ({ ...prev, [name]: value }))
    if (fieldErrors[name]) {
      setFieldErrors(prev => ({ ...prev, [name]: null }))
    }
    setAuthError(null)
  }

  const validateForm = () => {
    const errors = {}

    if (!form.email.trim()) {
      errors.email = 'Введите email'
    } else if (!/\S+@\S+\.\S+/.test(form.email)) {
      errors.email = 'Некорректный формат email'
    }

    if (!form.password) {
      errors.password = 'Введите пароль'
    } else if (mode === 'register' && form.password.length < 8) {
      errors.password = 'Пароль должен быть не менее 8 символов'
    }

    if (mode === 'register') {
      if (!form.firstName.trim()) errors.firstName = 'Введите имя'
      if (!form.lastName.trim()) errors.lastName = 'Введите фамилию'
      if (!form.phone.trim()) {
        errors.phone = 'Введите телефон'
      } else if (!/^\+?[\d\s\-\(\)]{10,}$/.test(form.phone)) {
        errors.phone = 'Некорректный формат телефона'
      }
    }

    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async () => {
    if (!validateForm()) return

    setLoading(true)
    setAuthError(null)
    setFieldErrors({})

    try {
      let res

      if (mode === 'login') {
        res = await authApi.login(form.email, form.password)
      } else {
        res = await authApi.register({
          login: form.email,
          email: form.email,
          phone: form.phone,
          password: form.password,
          first_name: form.firstName,
          last_name: form.lastName,
          middle_name: '',
          role: 'patient',
        })
      }

      if (res.data?.two_factor_required) {
        setPendingTwoFactor({
          challengeId: res.data.two_factor_challenge,
          expiresAt: res.data.two_factor_expires_at,
        })
        return
      }

      const fallbackName = [form.firstName, form.lastName].filter(Boolean).join(' ')
      persistAuthResponse(res.data, login, fallbackName)
      navigate('/patient')
    } catch (err) {
      const data = err?.response?.data
      const statusCode = err?.response?.status
      const rawMsg = data?.message || data?.error || ''

      if (statusCode === 409 || rawMsg?.toLowerCase?.().includes('already') || rawMsg?.toLowerCase?.().includes('exists')) {
        if (rawMsg?.toLowerCase?.().includes('email') || data?.field === 'email') {
          setFieldErrors(prev => ({ ...prev, email: 'Этот email уже зарегистрирован' }))
        } else if (rawMsg?.toLowerCase?.().includes('phone') || data?.field === 'phone') {
          setFieldErrors(prev => ({ ...prev, phone: 'Этот номер уже используется' }))
        } else {
          setAuthError('Пользователь с такими данными уже существует')
        }
      }
      else if (rawMsg === 'validation failed' || rawMsg === 'invalid request body') {
        const details = data?.details
        if (details && typeof details === 'string') {
          setAuthError(details)
        } else if (data?.errors && Array.isArray(data.errors)) {
          const fieldMap = {
            'email': 'email', 'login': 'email', 'phone': 'phone',
            'password': 'password', 'first_name': 'firstName', 'last_name': 'lastName'
          }
          const newFieldErrors = {}
          data.errors.forEach(e => {
            const fieldName = fieldMap[e.field]
            if (fieldName) newFieldErrors[fieldName] = e.message || 'Ошибка в поле'
          })
          if (Object.keys(newFieldErrors).length) {
            setFieldErrors(newFieldErrors)
          } else {
            setAuthError('Ошибка валидации: проверьте все поля')
          }
        } else {
          setAuthError('Ошибка валидации: проверьте все поля (пароль — минимум 8 символов)')
        }
      }
      else if (statusCode === 401 && mode === 'login') {
        setAuthError('Неверный email или пароль')
      }
      else {
        setAuthError(rawMsg || 'Произошла ошибка. Попробуйте ещё раз')
      }
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') handleSubmit()
  }

  if (pendingTwoFactor) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-black via-zinc-950 to-zinc-900 flex items-center justify-center p-4">
        <div className="luxury-card border border-amber-600/20 shadow-2xl shadow-black/50 w-full max-w-md p-8">
          <TwoFactorChallenge
            challenge={pendingTwoFactor}
            login={login}
            fallbackName={form.email}
            onBack={() => setPendingTwoFactor(null)}
            onComplete={() => navigate('/patient')}
          />
        </div>
      </div>
    )
  }


  return (
      <div className="min-h-screen bg-gradient-to-br from-black via-zinc-950 to-zinc-900 flex items-center justify-center p-4 relative overflow-hidden">
        {/* Декоративный фон */}
        <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-amber-600/10 rounded-full blur-3xl"></div>
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-amber-800/10 rounded-full blur-3xl"></div>
          <div className="absolute top-1/2 left-1/2 w-64 h-64 bg-violet-600/5 rounded-full blur-3xl"></div>
        </div>

        <div className="w-full max-w-md relative z-10">
          {/* Логотип */}
          <div className="flex items-center gap-3 mb-8 justify-center">
            <div className="w-14 h-14 bg-gradient-to-br from-amber-400 via-amber-500 to-amber-600 rounded-2xl flex items-center justify-center shadow-2xl shadow-amber-900/40 border border-amber-400/30">
              <svg width="28" height="28" viewBox="0 0 16 16" fill="none">
                <path d="M8 2v12M2 8h12" stroke="black" strokeWidth="2.5" strokeLinecap="round"/>
              </svg>
            </div>
            <div>
              <span className="text-3xl font-bold text-white tracking-tight block">MedBratishka</span>
              <p className="text-xs text-amber-500/90 text-center font-medium">Панель пациента</p>
            </div>
          </div>

          <div className="luxury-card border border-amber-600/20 shadow-2xl shadow-black/50">
            <div className="px-8 pt-8 pb-6 border-b border-amber-600/20">
              <h1 className="text-2xl font-bold text-gold mb-2">
                {mode === 'login' ? 'Добро пожаловать' : 'Регистрация'}
              </h1>
              <p className="text-sm text-zinc-500">Личный кабинет пациента</p>
            </div>

            <div className="p-8 flex flex-col gap-4">
              {mode === 'register' && (
                  <>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="flex flex-col gap-1.5">
                        <label className="text-xs text-zinc-400 font-medium">Имя *</label>
                        <input
                            className={`input-field ${fieldErrors.firstName ? 'border-red-500/50 focus:border-red-500 focus:ring-red-500/20' : ''}`}
                            name="firstName"
                            placeholder="Алёна"
                            value={form.firstName}
                            onChange={handleChange}
                            onKeyDown={handleKeyDown}
                        />
                        {fieldErrors.firstName && <p className="text-xs text-red-400">{fieldErrors.firstName}</p>}
                      </div>
                      <div className="flex flex-col gap-1.5">
                        <label className="text-xs text-zinc-400 font-medium">Фамилия *</label>
                        <input
                            className={`input-field ${fieldErrors.lastName ? 'border-red-500/50 focus:border-red-500 focus:ring-red-500/20' : ''}`}
                            name="lastName"
                            placeholder="Иванова"
                            value={form.lastName}
                            onChange={handleChange}
                            onKeyDown={handleKeyDown}
                        />
                        {fieldErrors.lastName && <p className="text-xs text-red-400">{fieldErrors.lastName}</p>}
                      </div>
                    </div>

                    <div className="flex flex-col gap-1.5">
                      <label className="text-xs text-zinc-400 font-medium">Телефон *</label>
                      <input
                          className={`input-field ${fieldErrors.phone ? 'border-red-500/50 focus:border-red-500 focus:ring-red-500/20' : ''}`}
                          name="phone"
                          type="tel"
                          placeholder="+79001234567"
                          value={form.phone}
                          onChange={handleChange}
                          onKeyDown={handleKeyDown}
                      />
                      {fieldErrors.phone && <p className="text-xs text-red-400">{fieldErrors.phone}</p>}
                    </div>
                  </>
              )}

              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400 font-medium">Email *</label>
                <input
                    className={`input-field ${fieldErrors.email ? 'border-red-500/50 focus:border-red-500 focus:ring-red-500/20' : ''}`}
                    name="email"
                    type="email"
                    placeholder="patient@mail.ru"
                    value={form.email}
                    onChange={handleChange}
                    onKeyDown={handleKeyDown}
                />
                {fieldErrors.email && <p className="text-xs text-red-400">{fieldErrors.email}</p>}
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-xs text-zinc-400 font-medium">Пароль *</label>
                <input
                    className={`input-field ${fieldErrors.password ? 'border-red-500/50 focus:border-red-500 focus:ring-red-500/20' : ''}`}
                    name="password"
                    type="password"
                    placeholder="••••••••"
                    value={form.password}
                    onChange={handleChange}
                    onKeyDown={handleKeyDown}
                />
                {mode === 'register' && form.password.length > 0 && form.password.length < 8 && !fieldErrors.password && (
                    <p className="text-xs text-amber-500">Минимум 8 символов ({form.password.length}/8)</p>
                )}
                {fieldErrors.password && <p className="text-xs text-red-400">{fieldErrors.password}</p>}
              </div>

              {mode === 'register' && (
                  <div className="bg-amber-950/20 border border-amber-600/20 rounded-xl p-4 mt-2">
                    <p className="text-sm text-amber-200/90">
                      <span className="font-semibold">💡 Важно:</span> <span className="text-amber-200/70">После регистрации вы сможете привязаться к врачу на странице «Мой врач», введя код приглашения.</span>
                    </p>
                  </div>
              )}

              {authError && (
                  <div className="bg-red-950/30 border border-red-900/50 rounded-xl px-4 py-3">
                    <p className="text-sm text-red-400 flex items-center gap-2">
                      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M8 2v12M8 12l4-4-4-4"/>
                      </svg>
                      {authError}
                    </p>
                  </div>
              )}

              <button
                  onClick={handleSubmit}
                  disabled={loading}
                  className="btn-primary w-full py-3 mt-2 disabled:opacity-50 disabled:cursor-not-allowed text-base font-semibold shadow-lg shadow-amber-900/30"
              >
                {loading ? (
                    <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"/>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                  </svg>
                  Загрузка...
                </span>
                ) : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
              </button>

              <div className="flex items-center gap-3 my-2">
                <div className="flex-1 h-px bg-zinc-800" />
                <span className="text-xs text-zinc-500 font-medium">
                {mode === 'login' ? 'нет аккаунта?' : 'есть аккаунт?'}
              </span>
                <div className="flex-1 h-px bg-zinc-800" />
              </div>

              <button
                  onClick={() => {
                    setMode(mode === 'login' ? 'register' : 'login')
                    setAuthError(null)
                    setFieldErrors({})
                  }}
                  className="btn-secondary w-full py-2.5 text-sm font-medium"
              >
                {mode === 'login' ? 'Зарегистрироваться' : 'Войти'}
              </button>
            </div>
          </div>

          <p className="text-center text-xs text-zinc-600 mt-6">
            © 2026 MedBratishka. Premium medical service
          </p>
        </div>
      </div>
  )
}
