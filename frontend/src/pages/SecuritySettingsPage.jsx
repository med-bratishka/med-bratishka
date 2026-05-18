import { useState } from 'react'
import { authApi } from '../api/index'

export default function SecuritySettingsPage() {
  const [setup, setSetup] = useState(null)
  const [code, setCode] = useState('')
  const [recoveryCodes, setRecoveryCodes] = useState([])
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const run = async (fn) => {
    setLoading(true)
    setError('')
    setMessage('')
    try {
      await fn()
    } catch (err) {
      if (err?.response?.status === 404) {
        setError('Запрос 2FA не дошел до backend: проверьте proxy /auth')
      } else {
        setError(err?.response?.data?.message || 'Не удалось выполнить действие')
      }
    } finally {
      setLoading(false)
    }
  }

  const startSetup = () => run(async () => {
    const res = await authApi.setup2FA()
    setSetup(res.data)
    setRecoveryCodes([])
  })

  const confirm = () => run(async () => {
    const res = await authApi.confirm2FA(code.trim())
    setRecoveryCodes(res.data?.recovery_codes || [])
    setSetup(null)
    setCode('')
    setMessage('Двухфакторная аутентификация включена')
  })

  const disable = () => run(async () => {
    await authApi.disable2FA(code.trim())
    setSetup(null)
    setRecoveryCodes([])
    setCode('')
    setMessage('Двухфакторная аутентификация выключена')
  })

  const regenerate = () => run(async () => {
    const res = await authApi.regenerateRecoveryCodes(code.trim())
    setRecoveryCodes(res.data?.recovery_codes || [])
    setCode('')
    setMessage('Новые recovery codes созданы')
  })

  return (
    <div className="min-h-screen bg-gradient-to-br from-black via-zinc-950 to-zinc-900 p-8">
      <div className="max-w-3xl mx-auto space-y-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Безопасность</h1>
          <p className="text-sm text-zinc-500 mt-1">Двухфакторная аутентификация и recovery codes</p>
        </div>

        <section className="luxury-card border border-amber-600/20 p-6 space-y-4">
          <div className="flex flex-col gap-2">
            <p className="text-sm font-semibold text-gray-200">2FA через приложение-аутентификатор</p>
            <p className="text-xs text-zinc-500">Секрет хранится на сервере только в зашифрованном виде.</p>
          </div>

          {!setup && (
            <button className="btn-primary w-fit" onClick={startSetup} disabled={loading}>
              Включить 2FA
            </button>
          )}

          {setup && (
            <div className="space-y-4">
              <div className="rounded-xl border border-zinc-800 bg-zinc-950/70 p-4 space-y-2">
                <p className="text-xs text-zinc-500">Secret</p>
                <p className="font-mono text-sm text-amber-300 break-all">{setup.secret}</p>
                <p className="text-xs text-zinc-500">OTPAuth URL</p>
                <p className="font-mono text-xs text-zinc-300 break-all">{setup.otpauth_url}</p>
              </div>
              <div className="flex gap-3">
                <input
                  className="input-field flex-1"
                  placeholder="Код из приложения"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                />
                <button className="btn-primary" onClick={confirm} disabled={loading || !code.trim()}>
                  Подтвердить
                </button>
              </div>
            </div>
          )}

          <div className="border-t border-zinc-800 pt-4 space-y-3">
            <p className="text-sm font-semibold text-gray-200">Управление</p>
            <div className="flex gap-3">
              <input
                className="input-field flex-1"
                placeholder="TOTP или recovery code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
              />
              <button className="btn-secondary" onClick={regenerate} disabled={loading || !code.trim()}>
                Новые коды
              </button>
              <button className="btn-secondary" onClick={disable} disabled={loading || !code.trim()}>
                Выключить
              </button>
            </div>
          </div>

          {recoveryCodes.length > 0 && (
            <div className="rounded-xl border border-amber-600/20 bg-amber-950/10 p-4">
              <p className="text-sm font-semibold text-amber-300 mb-3">Recovery codes показываются один раз</p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                {recoveryCodes.map((item) => (
                  <code key={item} className="text-xs text-gray-200 bg-black/30 rounded-lg px-3 py-2">{item}</code>
                ))}
              </div>
            </div>
          )}

          {message && <p className="text-sm text-emerald-400">{message}</p>}
          {error && <p className="text-sm text-red-400">{error}</p>}
        </section>
      </div>
    </div>
  )
}
