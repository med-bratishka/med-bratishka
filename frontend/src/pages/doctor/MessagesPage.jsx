import { useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useChatsWithMeta } from '../../hooks/useChatsWithMeta'

const AVATAR_COLORS = [
  ['#dbeafe', '#2563eb'],
  ['#ede9fe', '#7c3aed'],
  ['#d1fae5', '#059669'],
  ['#fef3c7', '#d97706'],
  ['#ffe4e6', '#e11d48'],
  ['#cffafe', '#0891b2'],
]

function Avatar({ name, id }) {
  const [bg, fg] = AVATAR_COLORS[(id ?? 0) % AVATAR_COLORS.length]
  const initials = name ? name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() : '?'
  return (
    <div style={{ background: bg, color: fg }}
      className="w-12 h-12 rounded-full flex items-center justify-center flex-shrink-0 text-sm font-semibold">
      {initials}
    </div>
  )
}

function formatTime(ts) {
  if (!ts) return ''
  const ms = ts > 1e10 ? ts : ts * 1000
  const d = new Date(ms)
  const now = new Date()
  if (d.toDateString() === now.toDateString())
    return d.toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return 'вчера'
  return d.toLocaleDateString('ru', { day: 'numeric', month: 'short' })
}

export default function MessagesPage() {
  const navigate = useNavigate()
  const { chats, meta, loading, markChatAsRead, reload } = useChatsWithMeta(8000)

  // Перезагружаем при возврате на страницу (после закрытия чата)
  useEffect(() => {
    const onVisible = () => { if (document.visibilityState === 'visible') reload() }
    document.addEventListener('visibilitychange', onVisible)
    // Также перезагружаем при фокусе окна
    window.addEventListener('focus', reload)
    return () => {
      document.removeEventListener('visibilitychange', onVisible)
      window.removeEventListener('focus', reload)
    }
  }, [reload])

  const open = (chat) => {
    markChatAsRead(chat.id)
    navigate('/doctor/chat', { state: { chatId: chat.id, chat } })
  }

  return (
    <div className="flex flex-col h-screen bg-white">
      <div className="px-4 py-3 border-b border-gray-100 flex-shrink-0">
        <h1 className="text-base font-semibold text-gray-900">Сообщения</h1>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          [1, 2, 3, 4].map(i => (
            <div key={i} className="flex items-center gap-3 px-4 py-3">
              <div className="w-12 h-12 rounded-full bg-gray-100 animate-pulse flex-shrink-0" />
              <div className="flex-1 space-y-2">
                <div className="flex justify-between">
                  <div className="h-3 bg-gray-100 rounded animate-pulse w-28" />
                  <div className="h-2.5 bg-gray-100 rounded animate-pulse w-8" />
                </div>
                <div className="h-2.5 bg-gray-100 rounded animate-pulse w-48" />
              </div>
            </div>
          ))
        ) : chats.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center px-8">
            <div className="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
              <svg width="28" height="28" viewBox="0 0 16 16" fill="none" stroke="#9ca3af" strokeWidth="1.2">
                <path d="M2 4.5C2 3.12 3.12 2 4.5 2h7C12.88 2 14 3.12 14 4.5v5c0 1.38-1.12 2.5-2.5 2.5H8l-3 2v-2H4.5C3.12 12 2 10.88 2 9.5v-5z"/>
              </svg>
            </div>
            <p className="text-sm font-medium text-gray-600 mb-1">Нет диалогов</p>
            <p className="text-xs text-gray-400 leading-relaxed">Сообщения появятся после привязки пациентов</p>
          </div>
        ) : (
          chats.map((chat) => {
            const m = meta[chat.id] ?? { unread: 0, lastMsg: '', lastTs: 0 }
            const name = chat.other_name || chat.other_login || `Пациент #${chat.patient_id}`
            const hasUnread = m.unread > 0
            return (
              <button key={chat.id} onClick={() => open(chat)}
                className="w-full flex items-center gap-3 px-4 py-3 hover:bg-gray-50 active:bg-gray-100 transition-colors text-left border-b border-gray-50">
                <div className="relative flex-shrink-0">
                  <Avatar name={name} id={chat.patient_id} />
                  {hasUnread && (
                    <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] rounded-full bg-brand-400 text-white text-[10px] font-medium flex items-center justify-center px-1">
                      {m.unread > 99 ? '99+' : m.unread}
                    </span>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-baseline justify-between mb-0.5">
                    <p className={`text-sm truncate ${hasUnread ? 'font-semibold text-gray-900' : 'font-medium text-gray-800'}`}>{name}</p>
                    <p className={`text-xs flex-shrink-0 ml-3 ${hasUnread ? 'text-brand-400 font-medium' : 'text-gray-400'}`}>{formatTime(m.lastTs)}</p>
                  </div>
                  <p className={`text-xs truncate ${hasUnread ? 'text-gray-700 font-medium' : 'text-gray-400'}`}>
                    {m.lastMsg || 'Нет сообщений'}
                  </p>
                </div>
              </button>
            )
          })
        )}
      </div>
    </div>
  )
}
