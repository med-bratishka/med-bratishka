import { useState } from 'react'

export default function InvitePage() {
  const [activeCode, setActiveCode] = useState('')
  const [copied, setCopied] = useState(false)

  const generateCode = () => {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
    const random = Array.from({ length: 5 }, () => chars[Math.floor(Math.random() * chars.length)]).join('')
    setActiveCode(`MED-${random}`)
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
          <button onClick={generateCode} className="btn-primary">Создать код</button>
        </div>
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
            </div>
        ) : (
            <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
              <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
                <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
                  <path d="M13 4H3a1 1 0 00-1 1v6a1 1 0 001 1h10a1 1 0 001-1V5a1 1 0 00-1-1z"/>
                  <path d="M3 4l5 4 5-4"/>
                </svg>
              </div>
              <p className="text-sm font-medium text-gray-600 mb-1">Нет активного кода</p>
              <p className="text-xs text-gray-400">Нажмите «Создать код» чтобы сгенерировать инвайт</p>
            </div>
        )}
      </div>
  )
}