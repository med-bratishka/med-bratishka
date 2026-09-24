import { useState, useEffect } from 'react'
import { doctorApi } from '../../api/index'

function generateCode() {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
  const random = Array.from({ length: 6 }, () => chars[Math.floor(Math.random() * chars.length)]).join('')
  return `DR-${random}`
}

export default function InvitePage() {
  const [code, setCode] = useState(null)
  const [copied, setCopied] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

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
      <div className="min-h-screen relative">
        {/* Фон */}
        <div className="fixed inset-0 bg-gradient-to-br from-black via-zinc-950 to-zinc-900">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-amber-600/10 rounded-full blur-3xl"></div>
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-emerald-800/5 rounded-full blur-3xl"></div>
        </div>

        {/* Контент */}
        <div className="relative z-10 p-8 max-w-2xl">
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-bold text-gold mb-2">Инвайт-код</h1>
              <p className="text-sm text-zinc-500">Подключение новых пациентов</p>
            </div>
            <button
                onClick={handleRotate}
                disabled={loading}
                className="btn-secondary flex items-center gap-2 disabled:opacity-50"
            >
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
                <path d="M13.5 8A5.5 5.5 0 112.9 5M2.5 2v3h3"/>
              </svg>
              {loading ? 'Обновление...' : 'Новый код'}
            </button>
          </div>

          {error && (
              <div className="luxury-card border-red-900/50 bg-red-950/20 mb-6">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-red-900/30 flex items-center justify-center flex-shrink-0">
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#ef4444" strokeWidth="2">
                      <path d="M8 2v12M8 12l4-4-4-4"/>
                    </svg>
                  </div>
                  <p className="text-sm text-red-400">{error}</p>
                </div>
              </div>
          )}

          <div className="luxury-card border border-amber-600/20 shadow-xl shadow-amber-900/10">
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-400 to-amber-600 flex items-center justify-center shadow-lg shadow-amber-900/30">
                <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="black" strokeWidth="2">
                  <path d="M13 4H3a1 1 0 00-1 1v6a1 1 0 001 1h10a1 1 0 001-1V5a1 1 0 00-1-1z"/>
                </svg>
              </div>
              <div>
                <p className="text-sm font-semibold text-gray-200">Код приглашения</p>
                <p className="text-xs text-zinc-500">Отправьте этот код пациенту</p>
              </div>
            </div>

            {loading && !code ? (
                <div className="flex items-center gap-4 p-6 bg-zinc-950/50 rounded-xl border border-zinc-800">
                  <div className="h-14 w-48 bg-zinc-800 rounded-lg animate-pulse" />
                  <div className="h-10 w-28 bg-zinc-800 rounded-lg animate-pulse" />
                </div>
            ) : (
                <div className="space-y-4">
                  <div className="flex items-center gap-4 p-6 bg-gradient-to-r from-zinc-950 to-zinc-900 rounded-xl border border-amber-600/30 shadow-inner">
                    <div className={`font-mono text-3xl font-bold tracking-widest transition-all duration-300 ${
                        loading ? 'opacity-40' : 'text-gold-animated'
                    }`}>
                      {code}
                    </div>
                    <button
                        onClick={handleCopy}
                        disabled={loading || copied}
                        className={`ml-auto px-4 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200 ${
                            copied
                                ? 'bg-emerald-600/20 text-emerald-400 border border-emerald-600/30'
                                : 'btn-primary'
                        }`}
                    >
                      {copied ? '✓ Скопировано' : 'Скопировать'}
                    </button>
                  </div>

                  <div className="flex items-start gap-3 p-4 bg-amber-950/20 rounded-lg border border-amber-600/20">
                    <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="#fbbf24" strokeWidth="1.5" className="flex-shrink-0 mt-0.5">
                      <path d="M8 1a7 7 0 100 14A7 7 0 008 1zM8 11a1 1 0 110-2 1 1 0 010 2zM7 4h2v6H7z"/>
                    </svg>
                    <div>
                      <p className="text-sm text-amber-200/90 font-medium mb-1">Важная информация</p>
                      <p className="text-xs text-amber-200/70 leading-relaxed">
                        Код одноразовый и сбрасывается при каждом создании нового.
                        Старый код перестаёт работать сразу после генерации нового.
                      </p>
                    </div>
                  </div>
                </div>
            )}
          </div>

          {/* Инструкция */}
          <div className="mt-6 luxury-card">
            <h3 className="text-sm font-semibold text-gray-300 mb-3 flex items-center gap-2">
              <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="#fbbf24" strokeWidth="1.5">
                <path d="M8 1a7 7 0 100 14A7 7 0 008 1zM8 11a1 1 0 110-2 1 1 0 010 2zM7 4h2v6H7z"/>
              </svg>
              Как это работает
            </h3>
            <ol className="space-y-2 text-sm text-zinc-400">
              <li className="flex items-start gap-2">
                <span className="text-amber-500 font-semibold">1.</span>
                <span>Скопируйте код и отправьте пациенту</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-amber-500 font-semibold">2.</span>
                <span>Пациент вводит код в разделе «Мой врач»</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-amber-500 font-semibold">3.</span>
                <span>После привязки пациент появится в вашем списке</span>
              </li>
            </ol>
          </div>
        </div>
      </div>
  )
}