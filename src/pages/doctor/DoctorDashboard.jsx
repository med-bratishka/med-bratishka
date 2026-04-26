import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { chatApi } from '../../api/index'

export default function DoctorDashboard() {
  const [chats, setChats] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const navigate = useNavigate()

  useEffect(() => {
    chatApi.getChats()
      .then((res) => {
        const data = res.data
        const list = data?.chats ?? data
        setChats(Array.isArray(list) ? list : [])
      })
      .catch((err) => setError(err?.response?.data?.message || 'Не удалось загрузить данные'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-medium text-gray-800">Мои пациенты</h1>
          <p className="text-sm text-gray-400 mt-0.5">{chats.length} на курировании</p>
        </div>
        <button onClick={() => navigate('/doctor/invite')} className="btn-primary flex items-center gap-1.5">
          <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
          Инвайт-код
        </button>
      </div>

      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: 'Всего пациентов', value: chats.length },
          { label: 'Активных чатов', value: chats.length },
          { label: 'Новых сообщений', value: 0 },
        ].map((m) => (
          <div key={m.label} className="bg-white rounded-xl border border-gray-100 p-4">
            <p className="text-xs text-gray-400 mb-1">{m.label}</p>
            <p className="text-2xl font-medium text-gray-800">{m.value}</p>
          </div>
        ))}
      </div>

      {loading ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex items-center justify-center">
          <p className="text-sm text-gray-400">Загрузка...</p>
        </div>
      ) : error ? (
        <div className="bg-red-50 border border-red-100 rounded-xl p-6 text-center">
          <p className="text-sm text-red-500">{error}</p>
        </div>
      ) : chats.length === 0 ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center justify-center text-center">
          <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5"><circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/></svg>
          </div>
          <p className="text-sm font-medium text-gray-600 mb-1">Пациентов пока нет</p>
          <p className="text-xs text-gray-400">Создайте инвайт-код и отправьте его пациенту</p>
        </div>
      ) : (
        <div className="bg-white border border-gray-100 rounded-xl divide-y divide-gray-50">
          {chats.map((chat) => (
            <button
              key={chat.id}
              onClick={() => navigate('/doctor/chat', { state: { chatId: chat.id, chat } })}
              className="w-full flex items-center gap-4 px-5 py-4 hover:bg-gray-50 transition-colors text-left"
            >
              <div className="w-9 h-9 rounded-full bg-brand-100 flex items-center justify-center flex-shrink-0">
                <span className="text-brand-600 text-sm font-medium">
                  {(chat.patient_name || chat.title || '?')[0].toUpperCase()}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-800 truncate">
                  {chat.patient_name || chat.title || `Пациент #${chat.id}`}
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
