import { useState } from 'react'
import { authApi } from '../../api/index'
import { persistAuthResponse } from '../../utils/authSession'

export default function TwoFactorChallenge({ challenge, onBack, onComplete, login, fallbackName }) {
  const [code, setCode] = useState('')
  const [recoveryMode, setRecoveryMode] = useState(false)
  const [rememberDevice, setRememberDevice] = useState(true)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async () => {
    const value = code.trim()
    if (!value) {
      setError(recoveryMode ? 'Введите recovery code' : 'Введите код из приложения')
      return
    }
    setLoading(true)
    setError('')
    try {
      const res = await authApi.verify2FA({
        challengeId: challenge.challengeId,
        code: value,
        recoveryCode: recoveryMode ? value : '',
        rememberDevice,
      })
      persistAuthResponse(res.data, login, fallbackName)
      onComplete(res.data.user)
    } catch (err) {
      const msg = err?.response?.data?.code === 'CHALLENGE_EXPIRED'
        ? 'Срок проверки истек, войдите заново'
        : 'Неверный код'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <div>
        <p className="text-sm font-semibold text-gray-200">Двухфакторная проверка</p>
        <p className="text-xs text-zinc-500 mt-1">Введите код из приложения-аутентификатора</p>
      </div>
      <input
        className="input-field w-full"
        value={code}
        onChange={(e) => setCode(e.target.value)}
        placeholder={recoveryMode ? 'ABCDE-FGHIJ-KLMNO' : '123456'}
        autoFocus
      />
      <label className="flex items-center gap-2 text-xs text-zinc-400">
        <input type="checkbox" checked={rememberDevice} onChange={(e) => setRememberDevice(e.target.checked)} />
        Запомнить это устройство
      </label>
      {error && <p className="text-xs text-red-400">{error}</p>}
      <button className="btn-primary w-full" onClick={submit} disabled={loading}>
        {loading ? 'Проверяем...' : 'Подтвердить'}
      </button>
      <div className="flex justify-between text-xs">
        <button className="text-amber-400" onClick={() => setRecoveryMode((v) => !v)}>
          {recoveryMode ? 'Использовать TOTP' : 'Recovery code'}
        </button>
        <button className="text-zinc-400" onClick={onBack}>Назад</button>
      </div>
    </div>
  )
}
