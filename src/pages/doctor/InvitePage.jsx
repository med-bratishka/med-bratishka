import { useState } from 'react'
import { doctorApi } from '../../api/index'

export default function InvitePage() {
  const [activeCode, setActiveCode] = useState('')
  const [copied, setCopied] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const generateCode = async () => {
    setLoading(true)
    setError(null)
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
    const random = Array.from({ length: 6 }, () => chars[Math.floor(Math.random() * chars.length)]).join('')
    const newCode = `DR-${random}`
    try {
      await doctorApi.setCode(newCode)
      setActiveCode(newCode)
    } catch (err) {
      const msg = err?.response?.data?.message || err?.response?.data?.error || ''
      if (msg === 'validation failed') {
        setError('Ошибка валидации кода на сервере')
      } else {
        setError(msg || 'Не удалось сохранить код')
      }
    } finally {
      setLoading(false)
    }
  }

  const copyCode = () => {
    navigator.clipboard.writeText(activeCode)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="p-6 max-w-xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-medium text-gray-800">Инвайт-коды</h1>
          <p className="text-sm text-gray-400 mt-0.5">Подключение новых пациентов</p>
        </div>
        <button onClick={generateCode} disabled={loading} className="btn-primary disabled:opacity-50">
          {loading ? 'Создание...' : 'Получить код'}
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-100 rounded-xl p-4 mb-4">
          <p className="text-sm text-red-500">{error}</p>
        </div>
      )}

      {activeCode ? (
        <div className="bg-white border border-gray-100 rounded-xl p-5">
          <p className="text-xs text-gray-400 mb-3">Отправьте этот код пациенту</p>
          <div className="flex items-center gap-4">
            <div className="font-mono text-2xl font-medium tracking-widest text-brand-600 bg-brand-50 px-5 py-3 rounded-lg">
              {activeCode}
            </div>
            <button onClick={copyCode} className="btn-secondary text-xs">
              {copied ? '✓ Скопировано' : 'Скопировать'}
            </button>
          </div>
          <p className="text-xs text-gray-400 mt-3">
            Пациент вводит этот код при регистрации или в настройках аккаунта
          </p>
        </div>
      ) : (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
          <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
              <rect x="2" y="4" width="12" height="9" rx="1.5"/>
              <path d="M5 4V3a3 3 0 016 0v1"/>
            </svg>
          </div>
          <p className="text-sm font-medium text-gray-600 mb-1">Нет активного кода</p>
          <p className="text-xs text-gray-400">Нажмите «Получить код» чтобы сгенерировать инвайт</p>
        </div>
      )}
    </div>
  )
}
