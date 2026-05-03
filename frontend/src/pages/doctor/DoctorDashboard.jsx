import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { chatApi, doctorApi } from '../../api/index'

const AVATAR_COLORS = [
  'bg-blue-100 text-blue-600',
  'bg-violet-100 text-violet-600',
  'bg-emerald-100 text-emerald-600',
  'bg-amber-100 text-amber-600',
  'bg-rose-100 text-rose-600',
  'bg-cyan-100 text-cyan-600',
]

const getLastSeen = (chatId) => parseInt(localStorage.getItem(`seen_${chatId}`) || '0', 10)
const setLastSeen = (chatId, msgId) => localStorage.setItem(`seen_${chatId}`, String(msgId))

function Avatar({ name, id }) {
  const color = AVATAR_COLORS[(id ?? 0) % AVATAR_COLORS.length]
  const initials = name ? name.split(' ').map((w) => w[0]).join('').slice(0, 2).toUpperCase() : '?'
  return (
    <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 text-sm font-semibold ${color}`}>
      {initials}
    </div>
  )
}

function IconButton({ onClick, title, children, danger }) {
  return (
    <button onClick={onClick} title={title}
      className={`w-8 h-8 rounded-lg flex items-center justify-center transition-colors ${danger ? 'text-gray-400 hover:text-red-500 hover:bg-red-50' : 'text-gray-400 hover:text-brand-600 hover:bg-brand-50'}`}>
      {children}
    </button>
  )
}

const IconChat = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M2 4.5C2 3.12 3.12 2 4.5 2h7C12.88 2 14 3.12 14 4.5v5c0 1.38-1.12 2.5-2.5 2.5H8l-3 2v-2H4.5C3.12 12 2 10.88 2 9.5v-5z"/>
  </svg>
)
const IconPrescription = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <rect x="3" y="1.5" width="10" height="13" rx="1.5"/>
    <path d="M6 5h4M6 8h4M6 11h2"/><path d="M9.5 11.5l2 2M11.5 11.5l-2 2"/>
  </svg>
)
const IconUnlink = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M6.5 9.5l-2 2a2.121 2.121 0 000 3 2.121 2.121 0 003 0l2-2"/>
    <path d="M9.5 6.5l2-2a2.121 2.121 0 000-3 2.121 2.121 0 00-3 0l-2 2"/>
    <path d="M2 2l12 12"/>
  </svg>
)

function PatientRow({ chat, unread, lastMsg, onChat, onUnlink }) {
  const name = chat.other_name || chat.other_login || `Пациент #${chat.patient_id}`
  const hasUnread = unread > 0
  return (
    <div className="flex items-center gap-3 px-5 py-3.5 hover:bg-gray-50 transition-colors group">
      <Avatar name={name} id={chat.patient_id} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium text-gray-800 truncate">{name}</p>
          {hasUnread && <span className="flex-shrink-0 w-2 h-2 rounded-full bg-brand-400" />}
        </div>
        <p className={`text-xs truncate mt-0.5 ${hasUnread ? 'text-brand-400 font-medium' : 'text-gray-400'}`}>
          {hasUnread ? `Новое сообщение: ${lastMsg}` : (lastMsg || 'Нет сообщений')}
        </p>
      </div>
      <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
        <IconButton onClick={() => onChat(chat)} title="Открыть чат"><IconChat /></IconButton>
        <IconButton onClick={() => onUnlink(chat)} title="Отвязать пациента" danger><IconUnlink /></IconButton>
      </div>
      <svg className="text-gray-300 flex-shrink-0" width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M6 4l4 4-4 4"/></svg>
    </div>
  )
}

export default function DoctorDashboard() {
  const [chats, setChats] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [unlinking, setUnlinking] = useState(null)
  const [chatMeta, setChatMeta] = useState({})
  const navigate = useNavigate()

  const loadChats = useCallback(async () => {
    try {
      const res = await chatApi.getChats()
      const arr = Array.isArray(res.data?.items ?? res.data?.chats ?? res.data)
        ? (res.data?.items ?? res.data?.chats ?? res.data)
        : []
      setChats(arr)

      const meta = {}
      await Promise.all(arr.map(async (chat) => {
        try {
          const msgRes = await chatApi.getMessages(chat.id)
          const msgs = msgRes.data?.items ?? msgRes.data?.messages ?? msgRes.data ?? []
          if (!Array.isArray(msgs) || msgs.length === 0) { meta[chat.id] = { unread: 0, lastMsg: '' }; return }
          const last = msgs[msgs.length - 1]
          const lastSeen = getLastSeen(chat.id)
          meta[chat.id] = {
            unread: msgs.filter(m => m.id > lastSeen).length,
            lastMsg: last.content || '',
          }
        } catch { meta[chat.id] = { unread: 0, lastMsg: '' } }
      }))
      setChatMeta(meta)
    } catch (err) {
      setError(err?.response?.data?.message || 'Не удалось загрузить данные')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadChats() }, [])
  useEffect(() => {
    const interval = setInterval(loadChats, 10000)
    return () => clearInterval(interval)
  }, [loadChats])

  const handleChat = (chat) => {
    // Помечаем сообщения прочитанными
    const meta = chatMeta[chat.id]
    if (meta?.unread > 0) {
      setChatMeta(prev => ({ ...prev, [chat.id]: { ...prev[chat.id], unread: 0 } }))
    }
    navigate('/doctor/chat', { state: { chatId: chat.id, chat } })
  }

  const handleUnlink = async (chat) => {
    const name = chat.other_name || chat.other_login || `Пациент #${chat.patient_id}`
    if (!window.confirm(`Отвязать ${name}? Чат будет закрыт.`)) return
    setUnlinking(chat.patient_id)
    try {
      await chatApi.closeChat(chat.id)
      await doctorApi.unlinkPatient(chat.patient_id)
      setChats(prev => prev.filter(c => c.id !== chat.id))
    } catch (err) {
      alert(err?.response?.data?.message || 'Не удалось отвязать пациента')
    } finally { setUnlinking(null) }
  }

  const totalUnread = Object.values(chatMeta).reduce((s, m) => s + (m.unread || 0), 0)

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-lg font-medium text-gray-800">Мои пациенты</h1>
          <p className="text-sm text-gray-400 mt-0.5">{loading ? 'Загрузка...' : `${chats.length} на курировании`}</p>
        </div>
        <button onClick={() => navigate('/doctor/invite')} className="btn-primary flex items-center gap-1.5">
          <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
          Инвайт-код
        </button>
      </div>

      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: 'Всего пациентов', value: loading ? '—' : chats.length },
          { label: 'Активных чатов', value: loading ? '—' : chats.length },
          { label: 'Новых сообщений', value: loading ? '—' : totalUnread },
        ].map((m) => (
          <div key={m.label} className="bg-white rounded-xl border border-gray-100 p-4">
            <p className="text-xs text-gray-400 mb-1">{m.label}</p>
            <p className="text-2xl font-medium text-gray-800">{m.value}</p>
          </div>
        ))}
      </div>

      {loading ? (
        <div className="bg-white border border-gray-100 rounded-xl divide-y divide-gray-50 overflow-hidden">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center gap-3 px-5 py-3.5">
              <div className="w-10 h-10 rounded-full bg-gray-100 animate-pulse flex-shrink-0" />
              <div className="flex-1 space-y-1.5">
                <div className="h-3 bg-gray-100 rounded animate-pulse w-36" />
                <div className="h-2.5 bg-gray-100 rounded animate-pulse w-24" />
              </div>
            </div>
          ))}
        </div>
      ) : error ? (
        <div className="bg-red-50 border border-red-100 rounded-xl p-6 text-center">
          <p className="text-sm text-red-500">{error}</p>
        </div>
      ) : chats.length === 0 ? (
        <div className="bg-white border border-gray-100 rounded-xl p-10 flex flex-col items-center text-center">
          <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
            <svg width="20" height="20" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.5">
              <circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/>
            </svg>
          </div>
          <p className="text-sm font-medium text-gray-600 mb-1">Пациентов пока нет</p>
          <p className="text-xs text-gray-400">Создайте инвайт-код и отправьте его пациенту</p>
        </div>
      ) : (
        <div className="bg-white border border-gray-100 rounded-xl divide-y divide-gray-50 overflow-hidden">
          {chats.map((chat) => (
            <PatientRow
              key={chat.id}
              chat={chat}
              unread={chatMeta[chat.id]?.unread ?? 0}
              lastMsg={chatMeta[chat.id]?.lastMsg ?? ''}
              onChat={handleChat}
              onUnlink={handleUnlink}
            />
          ))}
        </div>
      )}
    </div>
  )
}
