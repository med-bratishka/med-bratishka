import { useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useChatsWithMeta } from '../../hooks/useChatsWithMeta'

const AVATAR_COLORS = [
  ['bg-gradient-to-br from-amber-400 to-amber-600', 'text-black'],
  ['bg-gradient-to-br from-violet-400 to-violet-600', 'text-white'],
  ['bg-gradient-to-br from-emerald-400 to-emerald-600', 'text-black'],
  ['bg-gradient-to-br from-rose-400 to-rose-600', 'text-white'],
  ['bg-gradient-to-br from-cyan-400 to-cyan-600', 'text-black'],
]

function Avatar({ name, id }) {
  const [bg, fg] = AVATAR_COLORS[(id ?? 0) % AVATAR_COLORS.length]
  const initials = name ? name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() : '?'
  return (
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 text-sm font-bold shadow-lg ${bg} ${fg}`}>
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

  useEffect(() => {
    const onVisible = () => { if (document.visibilityState === 'visible') reload() }
    document.addEventListener('visibilitychange', onVisible)
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
      <div className="min-h-screen relative">
        {/* Фон */}
        <div className="fixed inset-0 bg-gradient-to-br from-black via-zinc-950 to-zinc-900">
          <div className="absolute top-0 right-1/4 w-96 h-96 bg-amber-600/10 rounded-full blur-3xl"></div>
          <div className="absolute bottom-1/4 left-1/4 w-96 h-96 bg-violet-800/5 rounded-full blur-3xl"></div>
        </div>

        {/* Контент */}
        <div className="relative z-10 flex flex-col h-screen">
          <div className="px-6 py-4 border-b border-amber-600/20 bg-zinc-900/50 backdrop-blur-md flex-shrink-0">
            <h1 className="text-2xl font-bold text-gold">Сообщения</h1>
          </div>

          <div className="flex-1 overflow-y-auto p-4">
            {loading ? (
                <div className="space-y-3">
                  {[1, 2, 3, 4].map(i => (
                      <div key={i} className="luxury-card animate-pulse">
                        <div className="flex items-center gap-4">
                          <div className="w-12 h-12 rounded-xl bg-zinc-800" />
                          <div className="flex-1 space-y-2">
                            <div className="flex justify-between">
                              <div className="h-4 bg-zinc-800 rounded w-32" />
                              <div className="h-3 bg-zinc-800 rounded w-12" />
                            </div>
                            <div className="h-3 bg-zinc-800 rounded w-48" />
                          </div>
                        </div>
                      </div>
                  ))}
                </div>
            ) : chats.length === 0 ? (
                <div className="luxury-card flex flex-col items-center justify-center text-center py-16">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-zinc-800 to-zinc-900 flex items-center justify-center mb-4 border border-zinc-700">
                    <svg width="28" height="28" viewBox="0 0 16 16" fill="none" stroke="#713f12" strokeWidth="1.2">
                      <path d="M2 4.5C2 3.12 3.12 2 4.5 2h7C12.88 2 14 3.12 14 4.5v5c0 1.38-1.12 2.5-2.5 2.5H8l-3 2v-2H4.5C3.12 12 2 10.88 2 9.5v-5z"/>
                    </svg>
                  </div>
                  <p className="text-base font-semibold text-gray-300 mb-2">Нет диалогов</p>
                  <p className="text-sm text-zinc-500">Сообщения появятся после привязки пациентов</p>
                </div>
            ) : (
                <div className="space-y-3">
                  {chats.map((chat) => {
                    const m = meta[chat.id] ?? { unread: 0, lastMsg: '', lastTs: 0 }
                    const name = chat.other_name || chat.other_login || `Пациент #${chat.patient_id}`
                    const hasUnread = m.unread > 0
                    return (
                        <button
                            key={chat.id}
                            onClick={() => open(chat)}
                            className={`w-full luxury-card text-left hover:border-amber-500/50 transition-all duration-300 group ${
                                hasUnread ? 'border-amber-600/40 bg-amber-950/10' : ''
                            }`}
                        >
                          <div className="flex items-center gap-4">
                            <div className="relative flex-shrink-0">
                              <Avatar name={name} id={chat.patient_id} />
                              {hasUnread && (
                                  <span className="absolute -top-1 -right-1 min-w-[20px] h-[20px] rounded-full bg-gradient-to-br from-amber-400 to-amber-600 text-black text-[11px] font-bold flex items-center justify-center px-1.5 shadow-lg shadow-amber-900/30">
                            {m.unread > 99 ? '99+' : m.unread}
                          </span>
                              )}
                            </div>
                            <div className="flex-1 min-w-0">
                              <div className="flex items-baseline justify-between mb-1">
                                <p className={`text-sm truncate ${hasUnread ? 'font-bold text-gray-200' : 'font-semibold text-gray-300'}`}>
                                  {name}
                                </p>
                                <p className={`text-xs flex-shrink-0 ml-3 ${hasUnread ? 'text-amber-400 font-semibold' : 'text-zinc-500'}`}>
                                  {formatTime(m.lastTs)}
                                </p>
                              </div>
                              <p className={`text-sm truncate ${hasUnread ? 'text-gray-300 font-medium' : 'text-zinc-500'}`}>
                                {m.lastMsg || 'Нет сообщений'}
                              </p>
                            </div>
                            <svg className="text-zinc-700 group-hover:text-amber-500/50 transition-colors flex-shrink-0" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                              <path d="M6 4l4 4-4 4"/>
                            </svg>
                          </div>
                        </button>
                    )
                  })}
                </div>
            )}
          </div>
        </div>
      </div>
  )
}