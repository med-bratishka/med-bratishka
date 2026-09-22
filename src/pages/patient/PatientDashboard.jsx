import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { chatApi, patientApi } from '../../api/index'

export default function PatientDashboard() {
  const [chats, setChats] = useState([])
  const [loading, setLoading] = useState(true)
  const [bindCode, setBindCode] = useState('')
  const [binding, setBinding] = useState(false)
  const [bindError, setBindError] = useState(null)
  const [bindSuccess, setBindSuccess] = useState(false)
  const navigate = useNavigate()

  const loadChats = () => {
    setLoading(true)
    chatApi.getChats()
      .then((res) => { const data = res.data; const list = data?.chats ?? data; setChats(Array.isArray(list) ? list : []) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadChats() }, [])

  const handleBind = async () => {
    if (!bindCode.trim()) return
    setBinding(true)
    setBindError(null)
    try {
      await patientApi.bindDoctor(bindCode.trim())
      setBindSuccess(true)
      setBindCode('')
      loadChats()
    } catch (err) {
      setBindError(err?.response?.data?.message || 'Неверный код врача')
    } finally {
      setBinding(false)
    }
  }

  return (
    <div className="p-6 max-w-xl">
      <div className="mb-6">
        <h1 className="text-lg font-medium text-gray-800">Мой врач</h1>
        <p className="text-sm text-gray-400 mt-0.5">Чаты и информация о лечении</p>
      </div>

      {/* Bind doctor section */}
      {chats.length === 0 && !loading && (
        <div className="bg-white border border-gray-100 rounded-xl p-5 mb-4">
          <p className="text-sm font-medium text-gray-700 mb-3">Привязаться к врачу</p>
          <p className="text-xs text-gray-400 mb-3">Введите код, который вам дал врач</p>
          <div className="flex gap-2">
            <input
              className="input-field flex-1 font-mono tracking-widest uppercase"
              placeholder="MED-XXXXX"
              value={bindCode}
              onChange={(e) => { setBindCode(e.target.value); setBindError(null); setBindSuccess(false) }}
              onKeyDown={(e) => e.key === 'Enter' && handleBind()}
            />
            <button onClick={handleBind} disabled={binding || !bindCode.trim()} className="btn-primary disabled:opacity-50">
              {binding ? '...' : 'Привязать'}
            </button>
          </div>
          {bindError && <p className="text-xs text-red-500 mt-2">{bindError}</p>}
          {bindSuccess && <p className="text-xs text-green-600 mt-2">✓ Успешно привязан к врачу!</p>}
        </div>
      )}

      {loading ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex items-center justify-center">
          <p className="text-sm text-gray-400">Загрузка...</p>
        </div>
      ) : chats.length === 0 ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
          <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
              <rect x="3" y="1" width="10" height="14" rx="1.5"/>
              <path d="M6 5h4M6 8h4M6 11h2"/>
            </svg>
          </div>
          <p className="text-sm font-medium text-gray-600 mb-1">Врач не подключён</p>
          <p className="text-xs text-gray-400">Введите код выше, чтобы привязаться к врачу</p>
        </div>
      ) : (
        <div className="bg-white border border-gray-100 rounded-xl divide-y divide-gray-50">
          {chats.map((chat) => (
            <button
              key={chat.id}
              onClick={() => navigate('/patient/chat', { state: { chatId: chat.id, chat } })}
              className="w-full flex items-center gap-4 px-5 py-4 hover:bg-gray-50 transition-colors text-left"
            >
              <div className="w-9 h-9 rounded-full bg-brand-100 flex items-center justify-center flex-shrink-0">
                <span className="text-brand-600 text-sm font-medium">
                  {(chat.doctor_name || chat.title || 'В')[0].toUpperCase()}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-800 truncate">
                  {chat.doctor_name || chat.title || `Врач #${chat.id}`}
                </p>
                <p className="text-xs text-gray-400 truncate mt-0.5">
                  {chat.last_message || 'Нет сообщений'}
                </p>
              </div>
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5"><path d="M6 4l4 4-4 4"/></svg>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
