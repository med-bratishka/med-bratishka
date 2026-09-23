import { useState, useEffect } from 'react'
import { doctorApi } from '../../api/index'

function generateCode() {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
  const random = Array.from({ length: 6 }, () => chars[Math.floor(Math.random() * chars.length)]).join('')
  return `DR-${random}`
}

export default function InvitePage() {
  const [code, setCode] = useState(null)       // null = ещё не загружен
  const [copied, setCopied] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // При каждом входе на страницу — сразу генерируем и сохраняем новый код
  useEffect(() => {
    let cancelled = false
    const registerNewCode = async () => {
      setLoading(true)
      setError(null)
      const newCode = generateCode()
      try {
        await doctorApi.setCode(newCode)
        if (!cancelled) setCode(newCode)
      } catch (err) {
        if (!cancelled) {
          setError(err?.response?.data?.message || err?.response?.data?.error || 'Не удалось создать код')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    registerNewCode()
    return () => { cancelled = true }
  }, [])

  const handleRotate = async () => {
    setLoading(true)
    setError(null)
    setCopied(false)
    const newCode = generateCode()
    try {
      await doctorApi.setCode(newCode)
      setCode(newCode)
    } catch (err) {
      setError(err?.response?.data?.message || err?.response?.data?.error || 'Не удалось обновить код')
    } finally {
      setLoading(false)
    }
  }

  const handleCopy = () => {
    if (!code) return
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="p-6 max-w-xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-medium text-gray-800">Инвайт-код</h1>
          <p className="text-sm text-gray-400 mt-0.5">Подключение новых пациентов</p>
        </div>
        <button onClick={handleRotate} disabled={loading} className="btn-secondary text-xs disabled:opacity-50 flex items-center gap-1.5">
          <svg width="12" height="12" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
            <path d="M13.5 8A5.5 5.5 0 112.9 5M2.5 2v3h3"/>
          </svg>
          {loading ? 'Обновление...' : 'Новый код'}
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-100 rounded-xl p-4 mb-4">
          <p className="text-sm text-red-500">{error}</p>
        </div>
      )}

      <div className="card">
        <p className="text-xs text-gray-400 mb-4">
          Отправьте этот код пациенту — он одноразовый и сбрасывается при каждом входе на эту страницу
        </p>

        {loading && !code ? (
          <div className="flex items-center gap-4">
            <div className="h-12 w-40 bg-gray-100 rounded-lg animate-pulse" />
            <div className="h-8 w-24 bg-gray-100 rounded-lg animate-pulse" />
          </div>
        ) : (
          <div className="flex items-center gap-4">
            <div className={`font-mono text-2xl font-medium tracking-widest px-5 py-3 rounded-lg transition-opacity ${loading ? 'opacity-40' : 'text-brand-600 bg-brand-50'}`}>
              {code}
            </div>
            <button onClick={handleCopy} disabled={loading} className="btn-secondary text-xs disabled:opacity-50">
              {copied ? '✓ Скопировано' : 'Скопировать'}
            </button>
          </div>
        )}

        <div className="mt-4 pt-4 border-t border-gray-50">
          <p className="text-xs text-gray-400">
            Старый код перестаёт работать сразу после генерации нового. Пациент вводит код в разделе «Мой врач».
          </p>
        </div>
      </div>
    </div>
  )
}
