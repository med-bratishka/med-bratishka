import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { chatApi, doctorApi } from '../../api/index'
import { useChatsWithMeta } from '../../hooks/useChatsWithMeta'

const AVATAR_COLORS = [
  'bg-gradient-to-br from-amber-400 to-amber-600 text-black',
  'bg-gradient-to-br from-violet-400 to-violet-600 text-white',
  'bg-gradient-to-br from-emerald-400 to-emerald-600 text-black',
  'bg-gradient-to-br from-rose-400 to-rose-600 text-white',
  'bg-gradient-to-br from-cyan-400 to-cyan-600 text-black',
]

function Avatar({ name, id }) {
  const color = AVATAR_COLORS[(id ?? 0) % AVATAR_COLORS.length]
  const initials = name ? name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() : '?'
  return (
      <div className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 text-sm font-bold shadow-lg ${color}`}>
        {initials}
      </div>
  )
}

function IconButton({ onClick, title, children, danger }) {
  return (
      <button onClick={onClick} title={title}
              className={`w-9 h-9 rounded-lg flex items-center justify-center transition-all duration-200 ${danger ? 'text-zinc-500 hover:text-red-400 hover:bg-red-500/10' : 'text-zinc-500 hover:text-amber-400 hover:bg-amber-500/10'}`}>
        {children}
      </button>
  )
}

const IconChat = () => (
    <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M2 4.5C2 3.12 3.12 2 4.5 2h7C12.88 2 14 3.12 14 4.5v5c0 1.38-1.12 2.5-2.5 2.5H8l-3 2v-2H4.5C3.12 12 2 10.88 2 9.5v-5z"/>
    </svg>
)

const IconUnlink = () => (
    <svg width="18" height="18" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6.5 9.5l-2 2a2.121 2.121 0 000 3 2.121 2.121 0 003 0l2-2"/>
      <path d="M9.5 6.5l2-2a2.121 2.121 0 000-3 2.121 2.121 0 00-3 0l-2 2"/>
      <path d="M2 2l12 12"/>
    </svg>
)

function PatientRow({ chat, unread, lastMsg, onChat, onUnlink }) {
  const name = chat.other_name || chat.other_login || `Пациент #${chat.patient_id}`
  const hasUnread = unread > 0
  return (
      <div className="group relative bg-zinc-900/60 backdrop-blur-sm border border-zinc-800 rounded-xl p-4 hover:border-amber-500/50 hover:shadow-xl hover:shadow-amber-900/10 transition-all duration-300">
        <div className="flex items-center gap-4">
          <Avatar name={name} id={chat.patient_id} />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <p className="text-sm font-semibold text-gray-200 truncate">{name}</p>
              {hasUnread && <span className="flex-shrink-0 w-2 h-2 rounded-full bg-amber-500 shadow-lg shadow-amber-500/50" />}
            </div>
            <p className={`text-xs truncate mt-0.5 ${hasUnread ? 'text-amber-400 font-medium' : 'text-zinc-500'}`}>
              {hasUnread ? `Новое: ${lastMsg}` : (lastMsg || 'Нет сообщений')}
            </p>
          </div>
          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
            <IconButton onClick={() => onChat(chat)} title="Открыть чат"><IconChat /></IconButton>
            <IconButton onClick={() => onUnlink(chat)} title="Отвязать пациента" danger><IconUnlink /></IconButton>
          </div>
          <svg className="text-zinc-700 flex-shrink-0 group-hover:text-amber-500/50 transition-colors" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M6 4l4 4-4 4"/></svg>
        </div>
      </div>
  )
}

export default function DoctorDashboard() {
  const navigate = useNavigate()
  const { chats, meta, loading, reload, markChatAsRead } = useChatsWithMeta(10000)
  const [unlinking, setUnlinking] = useState(null)

  const handleChat = (chat) => {
    markChatAsRead(chat.id)
    navigate('/doctor/chat', { state: { chatId: chat.id, chat } })
  }

  const handleUnlink = async (chat) => {
    const name = chat.other_name || chat.other_login || `Пациент #${chat.patient_id}`
    if (!window.confirm(`Отвязать ${name}? Чат будет закрыт.`)) return
    setUnlinking(chat.patient_id)
    try {
      await chatApi.closeChat(chat.id)
      await doctorApi.unlinkPatient(chat.patient_id)
      reload()
    } catch (err) {
      alert(err?.response?.data?.message || 'Не удалось отвязать пациента')
    } finally {
      setUnlinking(null)
    }
  }

  const chatsWithUnread = Object.values(meta).filter(m => m.unread > 0).length

  return (
      <div className="min-h-screen relative">
        {/* Фон */}
        <div className="fixed inset-0 bg-gradient-to-br from-black via-zinc-950 to-zinc-900">
          <div className="absolute top-0 right-1/4 w-96 h-96 bg-amber-600/10 rounded-full blur-3xl"></div>
          <div className="absolute bottom-1/4 left-1/4 w-96 h-96 bg-emerald-800/5 rounded-full blur-3xl"></div>
        </div>

        {/* Контент */}
        <div className="relative z-10 p-8 max-w-5xl">
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-bold text-gold mb-2">Мои пациенты</h1>
              <p className="text-sm text-zinc-500">{loading ? 'Загрузка...' : `${chats.length} на курировании`}</p>
            </div>
            <button onClick={() => navigate('/doctor/invite')} className="btn-primary flex items-center gap-2 shadow-lg shadow-amber-900/30">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><path d="M8 2v12M2 8h12"/></svg>
              Инвайт-код
            </button>
          </div>

          {/* Статистика */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
            {[
              { label: 'Всего пациентов', value: loading ? '—' : chats.length, icon: '👥', color: 'from-amber-500 to-amber-600' },
              { label: 'Активных чатов', value: loading ? '—' : chats.length, icon: '💬', color: 'from-emerald-500 to-emerald-600' },
              { label: 'Новых сообщений', value: loading ? '—' : chatsWithUnread, icon: '', color: 'from-violet-500 to-violet-600' },
            ].map((m) => (
                <div key={m.label} className="luxury-card">
                  <div className="flex items-center justify-between mb-3">
                    <p className="text-xs text-zinc-500">{m.label}</p>
                    <span className="text-xl">{m.icon}</span>
                  </div>
                  <p className="text-3xl font-bold text-gray-200">{m.value}</p>
                </div>
            ))}
          </div>

          {/* Список пациентов */}
          {loading ? (
              <div className="space-y-3">
                {[1, 2, 3].map(i => (
                    <div key={i} className="luxury-card animate-pulse">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-xl bg-zinc-800" />
                        <div className="flex-1 space-y-2">
                          <div className="h-4 bg-zinc-800 rounded w-32" />
                          <div className="h-3 bg-zinc-800 rounded w-24" />
                        </div>
                      </div>
                    </div>
                ))}
              </div>
          ) : chats.length === 0 ? (
              <div className="luxury-card flex flex-col items-center text-center py-16 border border-zinc-800">
                <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-zinc-800 to-zinc-900 flex items-center justify-center mb-4 border border-zinc-700">
                  <svg width="28" height="28" viewBox="0 0 16 16" fill="none" stroke="#713f12" strokeWidth="1.5">
                    <circle cx="8" cy="5" r="3"/><path d="M2 14c0-3.3 2.7-6 6-6s6 2.7 6 6"/>
                  </svg>
                </div>
                <p className="text-base font-semibold text-gray-300 mb-2">Пациентов пока нет</p>
                <p className="text-sm text-zinc-500 mb-4">Создайте инвайт-код и отправьте его пациенту</p>
                <button onClick={() => navigate('/doctor/invite')} className="btn-primary">
                  Создать инвайт-код
                </button>
              </div>
          ) : (
              <div className="space-y-3">
                {chats.map(chat => (
                    <PatientRow
                        key={chat.id}
                        chat={chat}
                        unread={meta[chat.id]?.unread ?? 0}
                        lastMsg={meta[chat.id]?.lastMsg ?? ''}
                        onChat={handleChat}
                        onUnlink={handleUnlink}
                    />
                ))}
              </div>
          )}
        </div>
      </div>
  )
}